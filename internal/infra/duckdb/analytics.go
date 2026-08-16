package duckdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/d3cie/dottie/internal/domain"
)

func (c *Client) InsertEvent(ctx context.Context, event domain.Event) error {
	_, err := c.SQL.ExecContext(ctx, `
		INSERT INTO events (id, website_id, visitor_id, session_id, event_name, path, hostname, title, referrer, country, city, device, browser, os, occurred_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.WebsiteID, event.VisitorID, event.SessionID, event.Name, event.Path, event.Hostname, event.Title, event.Referrer, event.Country, event.City, event.Device, event.Browser, event.OS, event.OccurredAt.UTC())
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (c *Client) Dashboard(ctx context.Context, websiteID, period string, now time.Time) (domain.Dashboard, error) {
	start, previousStart := periodBounds(period, now.UTC())
	current, err := c.aggregate(ctx, websiteID, start, now)
	if err != nil {
		return domain.Dashboard{}, err
	}
	previous, err := c.aggregate(ctx, websiteID, previousStart, start)
	if err != nil {
		return domain.Dashboard{}, err
	}
	visits, visitors, views, err := c.series(ctx, websiteID, start, now)
	if err != nil {
		return domain.Dashboard{}, err
	}
	var online int64
	if err := c.SQL.QueryRowContext(ctx, `SELECT COUNT(DISTINCT visitor_id) FROM events WHERE website_id = ? AND occurred_at >= ?`, websiteID, now.Add(-5*time.Minute)).Scan(&online); err != nil {
		return domain.Dashboard{}, fmt.Errorf("query online visitors: %w", err)
	}
	return domain.Dashboard{
		Visits:          KPI(current.visits, previous.visits),
		ViewsPerVisit:   KPI(divide(current.views, current.visits), divide(previous.views, previous.visits)),
		BounceRate:      KPI(current.bounceRate, previous.bounceRate),
		AverageDuration: KPI(current.duration, previous.duration),
		UniqueVisitors:  KPI(current.visitors, previous.visitors),
		PageViews:       KPI(current.views, previous.views),
		VisitSeries:     visits, VisitorSeries: visitors, PageViewSeries: views,
		OnlineVisitors: online, HasReceivedEvent: current.views > 0,
	}, nil
}

func (c *Client) Breakdowns(ctx context.Context, websiteID, period, dimension, search string, limit, offset int, now time.Time) ([]domain.BreakdownItem, int64, error) {
	columns := map[string]string{
		"pages": "path", "referrers": "referrer", "countries": "country", "cities": "city", "devices": "device", "browsers": "browser", "os": "os", "events": "event_name",
	}
	column, ok := columns[dimension]
	if !ok {
		return nil, 0, errors.New("unsupported breakdown dimension")
	}
	start, _ := periodBounds(period, now.UTC())
	query := fmt.Sprintf(`
		WITH grouped AS (
			SELECT COALESCE(NULLIF(%s, ''), 'unknown') AS key, COUNT(*)::BIGINT AS count
			FROM events WHERE website_id = ? AND occurred_at >= ? AND occurred_at < ? AND event_name = 'pageview'
			AND LOWER(COALESCE(%s, '')) LIKE LOWER(?) GROUP BY key
		), totals AS (SELECT COALESCE(SUM(count), 0)::BIGINT AS total FROM grouped)
		SELECT key, count, CASE WHEN total = 0 THEN 0 ELSE count * 100.0 / total END AS percent, total
		FROM grouped, totals ORDER BY count DESC, key ASC LIMIT ? OFFSET ?`, column, column)
	rows, err := c.SQL.QueryContext(ctx, query, websiteID, start, now, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s breakdown: %w", dimension, err)
	}
	defer rows.Close()
	items := make([]domain.BreakdownItem, 0)
	var total int64
	for rows.Next() {
		var item domain.BreakdownItem
		if err := rows.Scan(&item.Key, &item.Count, &item.Percent, &total); err != nil {
			return nil, 0, fmt.Errorf("scan breakdown: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate breakdown: %w", err)
	}
	return items, total, nil
}

func (c *Client) Visitors(ctx context.Context, websiteID, search string, limit, offset int) ([]domain.Visitor, int64, error) {
	filter := "%" + strings.ToLower(search) + "%"
	var total int64
	if err := c.SQL.QueryRowContext(ctx, `SELECT COUNT(DISTINCT visitor_id) FROM events WHERE website_id = ? AND LOWER(visitor_id) LIKE ?`, websiteID, filter).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count visitors: %w", err)
	}
	rows, err := c.SQL.QueryContext(ctx, `
		SELECT visitor_id, COALESCE(NULLIF(arg_max(country, occurred_at), ''), 'unknown'), COALESCE(NULLIF(arg_max(city, occurred_at), ''), 'unknown'),
			COALESCE(NULLIF(arg_max(referrer, occurred_at), ''), 'direct/unknown'), COALESCE(NULLIF(arg_max(device, occurred_at), ''), 'unknown'),
			COALESCE(NULLIF(arg_max(browser, occurred_at), ''), 'unknown'), COALESCE(NULLIF(arg_max(os, occurred_at), ''), 'unknown'),
			MIN(occurred_at), MAX(occurred_at), COUNT(*) FILTER (WHERE event_name = 'pageview')::BIGINT
		FROM events WHERE website_id = ? AND LOWER(visitor_id) LIKE ? GROUP BY visitor_id
		ORDER BY MAX(occurred_at) DESC LIMIT ? OFFSET ?`, websiteID, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query visitors: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Visitor, 0)
	for rows.Next() {
		var item domain.Visitor
		if err := rows.Scan(&item.ID, &item.Country, &item.City, &item.Referrer, &item.Device, &item.Browser, &item.OS, &item.FirstSeenAt, &item.LastActiveAt, &item.Views); err != nil {
			return nil, 0, fmt.Errorf("scan visitor: %w", err)
		}
		item.Name = "Anonymous " + strings.ToUpper(item.ID[:min(6, len(item.ID))])
		items = append(items, item)
	}
	return items, total, rows.Err()
}

type aggregate struct {
	visits     float64
	visitors   float64
	views      float64
	bounceRate float64
	duration   float64
}

func (c *Client) aggregate(ctx context.Context, websiteID string, start, end time.Time) (aggregate, error) {
	row := c.SQL.QueryRowContext(ctx, `
		WITH sessions AS (
			SELECT session_id, COUNT(*) FILTER (WHERE event_name = 'pageview') AS views,
				date_diff('second', MIN(occurred_at), MAX(occurred_at)) AS duration
			FROM events WHERE website_id = ? AND occurred_at >= ? AND occurred_at < ? GROUP BY session_id
		)
		SELECT COUNT(DISTINCT session_id)::DOUBLE,
			COUNT(DISTINCT visitor_id)::DOUBLE,
			COUNT(*) FILTER (WHERE event_name = 'pageview')::DOUBLE,
			COALESCE((SELECT AVG(CASE WHEN views <= 1 THEN 100.0 ELSE 0 END) FROM sessions), 0),
			COALESCE((SELECT AVG(duration) FROM sessions), 0)
		FROM events WHERE website_id = ? AND occurred_at >= ? AND occurred_at < ?`, websiteID, start, end, websiteID, start, end)
	var result aggregate
	if err := row.Scan(&result.visits, &result.visitors, &result.views, &result.bounceRate, &result.duration); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return aggregate{}, fmt.Errorf("query dashboard aggregate: %w", err)
	}
	return result, nil
}

func (c *Client) series(ctx context.Context, websiteID string, start, end time.Time) ([]domain.Point, []domain.Point, []domain.Point, error) {
	rows, err := c.SQL.QueryContext(ctx, `
		SELECT date_trunc('day', occurred_at) AS bucket, COUNT(DISTINCT session_id)::BIGINT,
			COUNT(DISTINCT visitor_id)::BIGINT, COUNT(*) FILTER (WHERE event_name = 'pageview')::BIGINT
		FROM events WHERE website_id = ? AND occurred_at >= ? AND occurred_at < ?
		GROUP BY bucket ORDER BY bucket`, websiteID, start, end)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query dashboard series: %w", err)
	}
	defer rows.Close()
	visits := make([]domain.Point, 0)
	visitors := make([]domain.Point, 0)
	views := make([]domain.Point, 0)
	for rows.Next() {
		var timestamp time.Time
		var visitCount, visitorCount, viewCount int64
		if err := rows.Scan(&timestamp, &visitCount, &visitorCount, &viewCount); err != nil {
			return nil, nil, nil, fmt.Errorf("scan dashboard series: %w", err)
		}
		visits = append(visits, domain.Point{Timestamp: timestamp, Value: visitCount})
		visitors = append(visitors, domain.Point{Timestamp: timestamp, Value: visitorCount})
		views = append(views, domain.Point{Timestamp: timestamp, Value: viewCount})
	}
	return visits, visitors, views, rows.Err()
}

func KPI(current, previous float64) domain.KPI {
	change := 0.0
	if previous > 0 {
		change = (current - previous) / previous * 100
	} else if current > 0 {
		change = 100
	}
	return domain.KPI{Value: round(current), Change: round(change)}
}

func periodBounds(period string, end time.Time) (time.Time, time.Time) {
	duration := 30 * 24 * time.Hour
	switch period {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "90d":
		duration = 90 * 24 * time.Hour
	}
	start := end.Add(-duration)
	return start, start.Add(-duration)
}

func divide(value, by float64) float64 {
	if by == 0 {
		return 0
	}
	return value / by
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}
