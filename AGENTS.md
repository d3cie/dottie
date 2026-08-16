# AGENTS

Read this file before making changes in this repository. Treat these instructions as required defaults unless the task says otherwise.

Dottie is an open-source, self-hosted web analytics product. It ships as one Go executable that serves the Huma API, the Svelte dashboard, and the browser tracking script. The repository uses Go, chi, Huma, sqlc, SQLite, DuckDB, Svelte 5, SvelteKit, Tailwind CSS v4, shadcn-svelte conventions, Orval, pnpm, and Turborepo.

## Repository structure

- `cmd/dottie` is the executable entry point and CLI wiring.
- `internal/app` composes runtime dependencies.
- `internal/domain` owns product models and service interfaces.
- `internal/httpapi` owns thin Huma operations and HTTP middleware.
- `internal/infra/sqlite` owns transactional persistence and generated sqlc code.
- `internal/infra/duckdb` owns analytical event storage and queries.
- `internal/config`, `internal/auth`, `internal/backup`, and `internal/doctor` own their named concerns.
- `internal/web` embeds generated frontend assets. Do not edit generated assets by hand.
- `web/apps/app` is the Svelte dashboard.
- `web/apps/client` is the small browser tracker served as `/tracker.js`.
- `web/packages/ui` owns reusable UI primitives.
- `web/packages/api-client` contains Orval-generated API functions and schemas.
- `migrations/sqlite` and `migrations/duckdb` own ordered database migrations.

## Working conventions

- Preserve unrelated worktree changes and keep changes scoped.
- Use `go fmt`, idiomatic error wrapping, explicit dependency construction, and small interfaces at consumer boundaries.
- Keep HTTP operations thin. Put business behavior in services or repositories.
- Use `context.Context` for request-scoped operations and never hide it in structs.
- Prefer standard-library packages unless a dependency materially improves correctness or maintainability.
- Use structured `log/slog` logging. Do not log secrets, passwords, session tokens, raw IP addresses, or request bodies.
- Use `pnpm`; do not introduce npm or Yarn lockfiles.
- Use Svelte 5 patterns and TypeScript inference. Prefer arrow functions for frontend handlers and helpers.
- Use components from `@dottie/ui` before creating app-local primitives.
- Preserve Peasy's compact visual language: lowercase navigation and page titles, restrained cards, clear density, and deliberate spacing. User-facing explanations and errors use normal sentence case.
- Keep chart components declarative and small. Data shaping belongs in API responses or focused helpers, not in chart lifecycle code.
- Make filters bookmarkable through URL search parameters.
- Never edit sqlc or Orval output by hand. Change its source and regenerate it.

## Generation

- `make generate` runs sqlc and Orval generation.
- The checked-in OpenAPI document is generated from the Huma operation registry.
- Generated files are committed so consumers can build from a source archive without extra global tools.
- A generation check must leave the worktree clean.

## Validation

- Run the narrowest relevant check first, then broaden it.
- Go changes: `go test ./...` and `go vet ./...`.
- Frontend changes: `pnpm --dir web check`, `pnpm --dir web lint`, and `pnpm --dir web build`.
- API changes: regenerate sqlc/Orval output and verify the worktree diff.
- UI changes: run Playwright smoke tests when the browser dependencies are available.
- Release changes: validate GitHub Actions with `actionlint` and test a production build.
- If a check cannot run, report exactly what was not verified and why.

## Database changes

- SQLite is for users, sessions, websites, and configuration-like transactional data.
- DuckDB is for event facts, visitor analytics, and aggregates.
- Add a new ordered migration; never rewrite a migration that may have shipped.
- Prefer sqlc queries for SQLite access. DuckDB queries live behind the analytics repository.
- Treat backup consistency and forward migration compatibility as part of schema design.

## Security and privacy

- Dottie is private by default: bind to loopback unless configured otherwise.
- Passwords use a slow password hash. Sessions use random opaque tokens and only their hashes are stored.
- Collector endpoints accept public website IDs, not administrative secrets.
- IP addresses may be used transiently but are not persisted.
- Do not add outbound telemetry. Any future telemetry must be opt-in.

## Commits

- Commit every coherent completed change. Do not leave finished work uncommitted.
- Use short imperative sentence-case subjects without Conventional Commit prefixes.
- Do not commit secrets, local databases, build caches, or unrelated changes.
