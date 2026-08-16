# Deployment

Dottie runs in the foreground and expects a service manager to restart it. It binds to loopback by default; put Caddy, Nginx, or Traefik in front of it for public HTTPS.

## Docker Compose

Set `DOTTIE_BASE_URL` in `docker-compose.yml`, then run:

```sh
docker compose up -d
```

The included port mapping only exposes Dottie to the host. A minimal Caddy configuration is:

```caddyfile
analytics.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

The tracking script is then available at `https://analytics.example.com/tracker.js`.

## Native service

```ini
[Unit]
Description=Dottie analytics
After=network-online.target

[Service]
ExecStart=/usr/local/bin/dottie start
Environment=DOTTIE_BASE_URL=https://analytics.example.com
User=dottie
Group=dottie
Restart=on-failure
ProtectSystem=strict
PrivateTmp=true
ReadWritePaths=/var/lib/dottie
Environment=DOTTIE_DATA_DIR=/var/lib/dottie

[Install]
WantedBy=multi-user.target
```

Run `dottie doctor` before starting the service. Stop the service before `dottie backup` or `dottie restore` so both embedded database files are at a consistent checkpoint.

