type TrackPayload = {
  website_id: string;
  visitor_id: string;
  session_id: string;
  name: "pageview" | "event";
  event?: string;
  path: string;
  hostname: string;
  title: string;
  referrer: string;
  timestamp: string;
};

export {};

declare global {
  interface Window {
    dottie?: {
      track: (event: string) => void;
    };
  }
}

const script = document.currentScript as HTMLScriptElement | null;
const websiteId = script?.dataset.websiteId;

if (script && websiteId && navigator.doNotTrack !== "1") {
  const collector = new URL("/api/v1/collect", script.src).toString();
  const visitorId = persistentID("dottie_visitor_id", localStorage);
  const sessionId = persistentID("dottie_session_id", sessionStorage);
  let lastPath = "";

  const send = (name: "pageview" | "event", event?: string) => {
    if (name === "pageview" && lastPath === location.href) return;
    if (name === "pageview") lastPath = location.href;
    const payload: TrackPayload = {
      website_id: websiteId,
      visitor_id: visitorId,
      session_id: sessionId,
      name,
      event,
      path: `${location.pathname}${location.search}`,
      hostname: location.hostname,
      title: document.title,
      referrer: document.referrer,
      timestamp: new Date().toISOString(),
    };
    void fetch(collector, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      keepalive: true,
      credentials: "omit",
    });
  };

  const navigation = () => queueMicrotask(() => send("pageview"));
  for (const method of ["pushState", "replaceState"] as const) {
    const original = history[method];
    history[method] = function (...args) {
      const result = original.apply(this, args);
      navigation();
      return result;
    };
  }
  addEventListener("popstate", navigation);
  window.dottie = { track: (event: string) => send("event", event) };
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", navigation, { once: true });
  } else {
    navigation();
  }
}

function persistentID(key: string, storage: Storage): string {
  const existing = storage.getItem(key);
  if (existing) return existing;
  const value = crypto.randomUUID();
  storage.setItem(key, value);
  return value;
}
