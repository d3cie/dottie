package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/d3cie/dottie/internal/domain"
	duckstore "github.com/d3cie/dottie/internal/infra/duckdb"
	sqlitestore "github.com/d3cie/dottie/internal/infra/sqlite"
)

func TestAnalyticsFlow(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()
	client.Jar, _ = cookiejar.New(nil)

	bootstrap := requestJSON(t, client, http.MethodGet, server.URL+"/api/v1/system/bootstrap", "", nil, nil)
	if bootstrap["setup_required"] != true {
		t.Fatalf("expected setup to be required, got %#v", bootstrap)
	}

	setup := requestJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/setup", "", map[string]any{
		"email": "admin@example.com", "password": "a sufficiently long password",
	}, nil)
	if setup["email"] != "admin@example.com" {
		t.Fatalf("unexpected setup response: %#v", setup)
	}

	websites := requestJSON(t, client, http.MethodPost, server.URL+"/api/v1/websites", sessionFromResponse(t, client, server.URL), map[string]any{
		"name": "Example", "domain": "example.com", "timezone": "Africa/Harare",
	}, nil)
	websiteID, ok := websites["id"].(string)
	if !ok || websiteID == "" {
		t.Fatalf("unexpected website response: %#v", websites)
	}

	collectorHeaders := map[string]string{"Origin": "https://example.com", "User-Agent": "Mozilla/5.0"}
	collected := requestJSON(t, client, http.MethodPost, server.URL+"/api/v1/collect", "", map[string]any{
		"website_id": websiteID, "visitor_id": "visitor-12345678", "session_id": "session-12345678",
		"name": "pageview", "path": "/docs", "hostname": "example.com", "title": "Docs",
		"referrer": "https://search.example/", "timestamp": time.Now().UTC(),
	}, collectorHeaders)
	if collected["accepted"] != true {
		t.Fatalf("expected event to be accepted: %#v", collected)
	}

	cookie := sessionFromResponse(t, client, server.URL)
	dashboard := requestJSON(t, client, http.MethodGet, server.URL+"/api/v1/websites/"+websiteID+"/dashboard?period=30d", cookie, nil, nil)
	visits := dashboard["visits"].(map[string]any)
	if visits["value"].(float64) != 1 || dashboard["has_received_event"] != true {
		t.Fatalf("unexpected dashboard response: %#v", dashboard)
	}

	visitors := requestJSON(t, client, http.MethodGet, server.URL+"/api/v1/websites/"+websiteID+"/visitors?limit=15&offset=0", cookie, nil, nil)
	if visitors["total"].(float64) != 1 {
		t.Fatalf("unexpected visitors response: %#v", visitors)
	}
}

func TestCollectorRejectsAnotherOrigin(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()
	client.Jar, _ = cookiejar.New(nil)
	requestJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/setup", "", map[string]any{
		"email": "admin@example.com", "password": "a sufficiently long password",
	}, nil)
	cookie := sessionFromResponse(t, client, server.URL)
	website := requestJSON(t, client, http.MethodPost, server.URL+"/api/v1/websites", cookie, map[string]any{
		"name": "Example", "domain": "example.com", "timezone": "UTC",
	}, nil)

	body, _ := json.Marshal(map[string]any{
		"website_id": website["id"], "visitor_id": "visitor-12345678", "session_id": "session-12345678",
		"name": "pageview", "path": "/", "hostname": "evil.example", "title": "Nope", "referrer": "",
	})
	request, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/collect", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.example")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		data, _ := io.ReadAll(response.Body)
		t.Fatalf("expected forbidden, got %d: %s", response.StatusCode, data)
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	directory := t.TempDir()
	ctx := context.Background()
	sqlite, err := sqlitestore.Open(ctx, filepath.Join(directory, "dottie.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	duckdb, err := duckstore.Open(ctx, filepath.Join(directory, "analytics.duckdb"))
	if err != nil {
		_ = sqlite.Close()
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := New(Dependencies{Domain: domain.NewService(sqlite.Queries), Analytics: duckdb, Logger: logger, Version: "test"})
	server := httptest.NewServer(api.Router)
	t.Cleanup(func() {
		server.Close()
		_ = duckdb.Close()
		_ = sqlite.Close()
	})
	return server
}

func sessionFromResponse(t *testing.T, client *http.Client, serverURL string) string {
	t.Helper()
	request, _ := http.NewRequest(http.MethodGet, serverURL+"/api/v1/auth/me", nil)
	for _, cookie := range client.Jar.Cookies(request.URL) {
		if cookie.Name == sessionCookie {
			return cookie.String()
		}
	}
	t.Fatal("session cookie not found")
	return ""
}

func requestJSON(t *testing.T, client *http.Client, method, target, cookie string, body any, headers map[string]string) map[string]any {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequest(method, target, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		request.Header.Set("Cookie", cookie)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		data, _ := io.ReadAll(response.Body)
		t.Fatalf("%s %s returned %d: %s", method, target, response.StatusCode, data)
	}
	var result map[string]any
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
