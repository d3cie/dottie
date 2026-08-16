FROM node:24-bookworm-slim AS web
WORKDIR /src
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml web/orval.config.ts web/svelte.config.js web/tsconfig.json web/vite.config.ts ./web/
COPY web/src ./web/src
RUN pnpm --dir web install --frozen-lockfile
RUN pnpm --dir web build

FROM node:24-bookworm-slim AS tracker
WORKDIR /src
RUN corepack enable
COPY tracker/package.json tracker/pnpm-lock.yaml tracker/tsconfig.json tracker/vite.config.ts ./tracker/
COPY tracker/src ./tracker/src
RUN pnpm --dir tracker install --frozen-lockfile
RUN pnpm --dir tracker build

FROM golang:1.25-bookworm AS go
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/build ./web/build
COPY --from=tracker /src/tracker/dist ./tracker/dist
RUN bash scripts/embed-web.sh
RUN go build -trimpath -ldflags "-s -w" -o /out/dottie ./cmd/dottie

FROM debian:bookworm-slim
RUN useradd --create-home --uid 10001 dottie
COPY --from=go /out/dottie /usr/local/bin/dottie
USER dottie
VOLUME ["/data"]
EXPOSE 8080
ENV DOTTIE_DATA_DIR=/data DOTTIE_ADDRESS=0.0.0.0:8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/usr/local/bin/dottie", "status"]
ENTRYPOINT ["/usr/local/bin/dottie"]
CMD ["start"]
