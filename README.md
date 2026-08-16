# Dottie

Dottie is a self-hosted, privacy-conscious web analytics server distributed as a single executable. It serves its dashboard, API, and browser tracker from the same process and stores all data on your machine.

> Dottie is usable but pre-release software. Back up your data before upgrading.

## Quick start

Install the latest release on Linux or macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/d3cie/dottie/main/scripts/install.sh | sh
```

```sh
dottie start
```

Open `http://127.0.0.1:8080`, create the local administrator, add a website, then install the tracker shown in the dashboard:

```html
<script defer src="https://analytics.example.com/tracker.js" data-website-id="YOUR_WEBSITE_ID"></script>
```

Because the script URL is hosted by Dottie, it automatically submits events to the same Dottie origin even when installed on another website.

For a public installation, set `DOTTIE_BASE_URL=https://analytics.example.com` and place Dottie behind an HTTPS reverse proxy. See [docs/deployment.md](docs/deployment.md) for Docker Compose and systemd examples, and [docs/configuration.md](docs/configuration.md) for all settings.

## CLI

```text
dottie start
dottie status
dottie doctor
dottie backup
dottie restore
dottie admin reset-password
dottie config
```

## Development

Requirements: Go 1.25+, Node.js 22+, pnpm 10+, and sqlc.

```sh
make setup
make generate
make dev
make test
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the repository workflow, [docs/architecture.md](docs/architecture.md) for the design, and [docs/releases.md](docs/releases.md) for the release process.

## License

Dottie is licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE).
