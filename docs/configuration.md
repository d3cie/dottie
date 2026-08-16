# Configuration

Dottie loads JSON configuration from its data directory, then applies environment variables, then applies flags passed to `dottie start`. Run `dottie config` to print the effective values.

| JSON key | Environment variable | Default | Purpose |
| --- | --- | --- | --- |
| `data_dir` | `DOTTIE_DATA_DIR` | OS application data directory | SQLite, DuckDB, and configuration files |
| `address` | `DOTTIE_ADDRESS` | `127.0.0.1:8080` | HTTP listen address |
| `base_url` | `DOTTIE_BASE_URL` | Derived from the listen address | Public HTTPS origin used for links and cookies |
| `log_level` | `DOTTIE_LOG_LEVEL` | `info` | `info` or `debug` structured logs |

`PORT` is also accepted and becomes `0.0.0.0:$PORT` when `DOTTIE_ADDRESS` is unset. Use `--config /path/to/config.json` to load a specific file. The `start` command additionally accepts `--address`, `--data-dir`, and `--base-url`.

Set `DOTTIE_BASE_URL` for every public deployment. It makes session cookies secure and ensures the dashboard shows the correct same-origin tracker URL.
