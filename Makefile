.PHONY: setup generate generate-api build-web build dev test check clean

setup:
	go mod download
	pnpm --dir web install --frozen-lockfile
	pnpm --dir tracker install --frozen-lockfile

generate: generate-api
	sqlc generate

generate-api:
	go run ./cmd/dottie openapi --output web/openapi.yaml
	pnpm --dir web generate

build-web:
	pnpm --dir web build
	pnpm --dir tracker build
	bash scripts/embed-web.sh

build: build-web
	go build -o dist/dottie ./cmd/dottie

dev:
	@echo "Run these in separate terminals:"
	@echo "  go run ./cmd/dottie start"
	@echo "  pnpm --dir web dev"
	@echo "  pnpm --dir tracker dev"

test:
	go test ./...
	pnpm --dir web check
	pnpm --dir tracker check

check: generate
	go test ./...
	go vet ./...
	pnpm --dir web check
	pnpm --dir tracker check
	pnpm --dir web lint
	git diff --exit-code

clean:
	go clean
	find internal/web/dist -mindepth 1 -maxdepth 1 -delete
	find dist -mindepth 1 -maxdepth 1 -delete 2>/dev/null || true
