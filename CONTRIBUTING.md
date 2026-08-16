# Contributing to Dottie

Thank you for helping improve Dottie.

## Setup

```sh
git clone https://github.com/d3cie/dottie.git
cd dottie
make setup
make generate
make test
```

Run the backend and frontend development servers with `make dev`. The frontend proxies `/api`, `/tracker.js`, and `/openapi.json` to the Go server.

## Pull requests

- Keep changes focused and include tests for behavior changes.
- Run `make check` before opening a pull request.
- Do not edit generated sqlc, Orval, or embedded asset output by hand.
- Use short imperative commit subjects without Conventional Commit prefixes.

Please report security issues privately through GitHub's security advisory feature rather than a public issue.

