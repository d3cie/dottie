package domain

import "time"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Website struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
}

type Event struct {
	ID         string
	WebsiteID  string
	VisitorID  string
	SessionID  string
	Name       string
	Path       string
	Hostname   string
	Title      string
	Referrer   string
	Country    string
	City       string
	Device     string
	Browser    string
	OS         string
	OccurredAt time.Time
}

type KPI struct {
	Value  float64 `json:"value"`
	Change float64 `json:"change"`
}

type Point struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int64     `json:"value"`
}

type Dashboard struct {
	Visits           KPI     `json:"visits"`
	ViewsPerVisit    KPI     `json:"views_per_visit"`
	BounceRate       KPI     `json:"bounce_rate"`
	AverageDuration  KPI     `json:"average_duration"`
	UniqueVisitors   KPI     `json:"unique_visitors"`
	PageViews        KPI     `json:"page_views"`
	VisitSeries      []Point `json:"visit_series"`
	VisitorSeries    []Point `json:"visitor_series"`
	PageViewSeries   []Point `json:"page_view_series"`
	OnlineVisitors   int64   `json:"online_visitors"`
	HasReceivedEvent bool    `json:"has_received_event"`
}

type BreakdownItem struct {
	Key     string  `json:"key"`
	Count   int64   `json:"count"`
	Percent float64 `json:"percent"`
}

type Visitor struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	City         string    `json:"city"`
	Referrer     string    `json:"referrer"`
	Device       string    `json:"device"`
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	FirstSeenAt  time.Time `json:"first_seen_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	Views        int64     `json:"views"`
}
