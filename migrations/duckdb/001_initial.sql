CREATE TABLE IF NOT EXISTS events (
  id VARCHAR PRIMARY KEY,
  website_id VARCHAR NOT NULL,
  visitor_id VARCHAR NOT NULL,
  session_id VARCHAR NOT NULL,
  event_name VARCHAR NOT NULL,
  path VARCHAR NOT NULL,
  hostname VARCHAR NOT NULL,
  title VARCHAR NOT NULL,
  referrer VARCHAR NOT NULL,
  country VARCHAR NOT NULL,
  city VARCHAR NOT NULL,
  device VARCHAR NOT NULL,
  browser VARCHAR NOT NULL,
  os VARCHAR NOT NULL,
  occurred_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS events_website_time_idx ON events(website_id, occurred_at);
CREATE INDEX IF NOT EXISTS events_website_visitor_idx ON events(website_id, visitor_id);

