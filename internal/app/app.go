package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/d3cie/dottie/internal/config"
	"github.com/d3cie/dottie/internal/domain"
	"github.com/d3cie/dottie/internal/httpapi"
	duckstore "github.com/d3cie/dottie/internal/infra/duckdb"
	sqlitestore "github.com/d3cie/dottie/internal/infra/sqlite"
	webassets "github.com/d3cie/dottie/internal/web"
)

type App struct {
	server    *http.Server
	sqlite    *sqlitestore.Client
	duckdb    *duckstore.Client
	logger    *slog.Logger
	publicURL string
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger, version string) (*App, error) {
	if err := config.Save(cfg); err != nil {
		return nil, err
	}
	sqliteClient, err := sqlitestore.Open(ctx, cfg.SQLitePath())
	if err != nil {
		return nil, err
	}
	duckClient, err := duckstore.Open(ctx, cfg.DuckDBPath())
	if err != nil {
		_ = sqliteClient.Close()
		return nil, err
	}

	api := httpapi.New(httpapi.Dependencies{
		Domain: domain.NewService(sqliteClient.Queries), Analytics: duckClient, Logger: logger,
		SecureCookies: strings.HasPrefix(cfg.PublicURL(), "https://"), Version: version,
	})
	api.Router.NotFound(spaHandler())
	return &App{
		server: &http.Server{
			Addr: cfg.Address, Handler: api.Router, ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 2 * time.Minute,
		},
		sqlite: sqliteClient, duckdb: duckClient, logger: logger, publicURL: cfg.PublicURL(),
	}, nil
}

func (a *App) Start() error {
	a.logger.Info("dottie started", "address", a.server.Addr, "url", a.publicURL)
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve HTTP: %w", err)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("stopping dottie")
	serverErr := a.server.Shutdown(ctx)
	duckErr := a.duckdb.Close()
	sqliteErr := a.sqlite.Close()
	return errors.Join(serverErr, duckErr, sqliteErr)
}

func spaHandler() http.HandlerFunc {
	dist, err := fs.Sub(webassets.Assets, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(dist))
	return func(writer http.ResponseWriter, request *http.Request) {
		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested == "." || requested == "" {
			requested = "index.html"
		}
		if info, statErr := fs.Stat(dist, requested); statErr == nil && !info.IsDir() {
			if contentType := mime.TypeByExtension(path.Ext(requested)); contentType != "" {
				writer.Header().Set("Content-Type", contentType)
			}
			if strings.Contains(requested, "/_app/immutable/") || requested == "tracker.js" {
				writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			files.ServeHTTP(writer, request)
			return
		}
		index, readErr := fs.ReadFile(dist, "index.html")
		if readErr != nil {
			http.Error(writer, "Dottie frontend has not been built. Run make build-web.", http.StatusServiceUnavailable)
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Header().Set("Cache-Control", "no-cache")
		_, _ = writer.Write(index)
	}
}
