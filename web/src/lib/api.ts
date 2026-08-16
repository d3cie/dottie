import {
  createWebsite as generatedCreateWebsite,
  getBootstrapStatus,
  getBreakdowns,
  getCurrentUser,
  getDashboard,
  listVisitors,
  listWebsites,
  login as generatedLogin,
  logout as generatedLogout,
  setupAdmin,
} from '$lib/api-client/generated/dottie';
import type {
  GetBreakdownsDimension,
  GetBreakdownsPeriod,
  GetDashboardPeriod,
} from '$lib/api-client/generated/models';

export type User = { id: string; email: string };
export type Website = {
  id: string;
  name: string;
  domain: string;
  timezone: string;
  created_at: string;
};
export type KPI = { value: number; change: number };
export type Point = { timestamp: string; value: number };
export type Dashboard = {
  visits: KPI;
  views_per_visit: KPI;
  bounce_rate: KPI;
  average_duration: KPI;
  unique_visitors: KPI;
  page_views: KPI;
  visit_series: Point[];
  visitor_series: Point[];
  page_view_series: Point[];
  online_visitors: number;
  has_received_event: boolean;
};
export type Breakdown = { key: string; count: number; percent: number };
export type Visitor = {
  id: string;
  name: string;
  country: string;
  city: string;
  referrer: string;
  device: string;
  browser: string;
  os: string;
  first_seen_at: string;
  last_active_at: string;
  views: number;
};

const options: RequestInit = { credentials: 'include' };

export const api = {
  bootstrap: () =>
    getBootstrapStatus(options) as Promise<{ setup_required: boolean }>,
  setup: (email: string, password: string) =>
    setupAdmin({ email, password }, options) as Promise<User>,
  login: (email: string, password: string) =>
    generatedLogin({ email, password }, options) as Promise<User>,
  logout: () => generatedLogout(options),
  me: () => getCurrentUser(options) as Promise<User>,
  websites: () => listWebsites(options) as Promise<{ websites: Website[] }>,
  createWebsite: (name: string, domain: string, timezone: string) =>
    generatedCreateWebsite(
      { name, domain, timezone },
      options,
    ) as Promise<Website>,
  dashboard: (websiteId: string, period: string) =>
    getDashboard(
      websiteId,
      { period: period as GetDashboardPeriod },
      options,
    ) as Promise<Dashboard>,
  breakdown: (websiteId: string, period: string, dimension: string) =>
    getBreakdowns(
      websiteId,
      {
        period: period as GetBreakdownsPeriod,
        dimension: dimension as GetBreakdownsDimension,
        limit: 9,
        offset: 0,
      },
      options,
    ) as Promise<{ items: Breakdown[]; total: number }>,
  visitors: (websiteId: string, page: number, search: string) =>
    listVisitors(
      websiteId,
      { limit: 15, offset: (page - 1) * 15, search: search || undefined },
      options,
    ) as Promise<{ visitors: Visitor[]; total: number }>,
};

export const errorMessage = (error: unknown) => {
  if (error instanceof Error) return error.message;
  return 'An unexpected error occurred.';
};
