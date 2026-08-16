# Changelog

All notable changes to Dottie will be documented here. Dottie follows Semantic Versioning once the first version is tagged.

## Unreleased

### Added

- Single-process Go server with an embedded Svelte dashboard and browser tracker.
- Local administrator authentication, multiple websites, DuckDB analytics, SQLite configuration data, backups, diagnostics, and generated API clients.
- Dashboard and visitors pages based on the Peasy visual language.
- Current shadcn-svelte primitives for buttons, inputs, labels, selects, cards, tables, avatars, skeletons, and keyboard hints.

### Changed

- Rebuilt the dashboard and visitors interface around Peasy's page container, responsive sidebar, spacing, elevated surfaces, KPI strip, chart cards, count lists, and dense visitor table.
- Flattened `web` into one SvelteKit application and moved the independent browser script into `tracker`, removing Turborepo and all frontend workspace packages.
- Matched Peasy's typography by loading Nunito at weights 200–900 and Lora Variable for serif headings, and moved the dashboard controls onto shadcn-svelte without losing Peasy's sizing and green treatment.
