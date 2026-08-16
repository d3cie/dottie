# Architecture

Dottie is a single-node, self-hosted analytics service. One Go process owns the HTTP server and both embedded databases.

```text
tracked website -> /tracker.js -> POST /api/v1/collect -> DuckDB
                                                |
browser dashboard -> Huma API -> services -> SQLite + DuckDB
```

SQLite contains transactional state such as the administrator, sessions, and websites. sqlc generates its query layer. DuckDB contains immutable event facts and serves analytical queries. Administrative API routes require a secure session cookie; collection uses a public website ID and validates the request origin when the website has an origin configured.

The Svelte app and tracker are built before Go release builds and embedded into the executable. `/tracker.js` derives its collection endpoint from its own script URL, so it can be installed on any website while posting back to the Dottie host.

## Runtime files

The default data directory follows the operating system convention. It contains:

```text
config.json
dottie.sqlite
analytics.duckdb
```

The process is intended to run in the foreground behind a service manager and reverse proxy. It binds to `127.0.0.1:8080` by default.

