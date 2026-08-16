package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/d3cie/dottie/internal/app"
	"github.com/d3cie/dottie/internal/backup"
	"github.com/d3cie/dottie/internal/config"
	"github.com/d3cie/dottie/internal/domain"
	"github.com/d3cie/dottie/internal/httpapi"
	duckstore "github.com/d3cie/dottie/internal/infra/duckdb"
	sqlitestore "github.com/d3cie/dottie/internal/infra/sqlite"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var configPath string
	root := &cobra.Command{
		Use:           "dottie",
		Short:         "Self-hosted web analytics",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to a configuration file")
	load := func() (config.Config, error) { return config.Load(configPath) }
	root.AddCommand(startCommand(load), statusCommand(load), doctorCommand(load), configCommand(load), backupCommand(load), restoreCommand(load), adminCommand(load), openAPICommand())
	return root
}

type configLoader func() (config.Config, error)

func startCommand(load configLoader) *cobra.Command {
	var address, dataDir, baseURL string
	command := &cobra.Command{
		Use: "start", Short: "Start the Dottie server",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := load()
			if err != nil {
				return err
			}
			if address != "" {
				cfg.Address = address
			}
			if dataDir != "" {
				cfg.DataDir = dataDir
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}
			logger := newLogger(cfg.LogLevel)
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			application, err := app.New(ctx, cfg, logger, version)
			if err != nil {
				return err
			}
			errCh := make(chan error, 1)
			go func() { errCh <- application.Start() }()
			select {
			case err := <-errCh:
				return err
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				return application.Shutdown(shutdownCtx)
			}
		},
	}
	command.Flags().StringVar(&address, "address", "", "HTTP listen address")
	command.Flags().StringVar(&dataDir, "data-dir", "", "directory for Dottie data")
	command.Flags().StringVar(&baseURL, "base-url", "", "public URL used for links and secure cookies")
	return command
}

func statusCommand(load configLoader) *cobra.Command {
	var jsonOutput bool
	command := &cobra.Command{Use: "status", Short: "Show whether Dottie is running", RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		client := &http.Client{Timeout: 2 * time.Second}
		response, err := client.Get(cfg.PublicURL() + "/health")
		running := err == nil && response.StatusCode == http.StatusOK
		if response != nil {
			_ = response.Body.Close()
		}
		value := map[string]any{"running": running, "url": cfg.PublicURL(), "data_dir": cfg.DataDir}
		if jsonOutput {
			return printJSON(value)
		}
		if running {
			fmt.Printf("Dottie is running at %s\n", cfg.PublicURL())
			return nil
		}
		fmt.Printf("Dottie is not responding at %s\n", cfg.PublicURL())
		return errors.New("server is not running")
	}}
	command.Flags().BoolVar(&jsonOutput, "json", false, "print machine-readable JSON")
	return command
}

func doctorCommand(load configLoader) *cobra.Command {
	var jsonOutput bool
	command := &cobra.Command{Use: "doctor", Short: "Check Dottie's configuration and databases", RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		checks := []map[string]any{{"name": "configuration", "ok": true}, {"name": "data directory", "ok": true}}
		if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
			checks[1]["ok"] = false
			checks[1]["error"] = err.Error()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		sqlite, sqliteErr := sqlitestore.Open(ctx, cfg.SQLitePath())
		checks = append(checks, checkResult("SQLite", sqliteErr))
		if sqlite != nil {
			_ = sqlite.Close()
		}
		duckdb, duckErr := duckstore.Open(ctx, cfg.DuckDBPath())
		checks = append(checks, checkResult("DuckDB", duckErr))
		if duckdb != nil {
			_ = duckdb.Close()
		}
		if jsonOutput {
			_ = printJSON(map[string]any{"checks": checks})
			return errors.Join(sqliteErr, duckErr)
		}
		failed := false
		for _, check := range checks {
			mark := "✓"
			if check["ok"] != true {
				mark = "✗"
				failed = true
			}
			fmt.Printf("%s %s", mark, check["name"])
			if check["error"] != nil {
				fmt.Printf(": %s", check["error"])
			}
			fmt.Println()
		}
		if failed {
			return errors.New("one or more checks failed")
		}
		return nil
	}}
	command.Flags().BoolVar(&jsonOutput, "json", false, "print machine-readable JSON")
	return command
}

func configCommand(load configLoader) *cobra.Command {
	return &cobra.Command{Use: "config", Short: "Print the effective configuration", RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		return printJSON(cfg)
	}}
}

func backupCommand(load configLoader) *cobra.Command {
	var output string
	command := &cobra.Command{Use: "backup", Short: "Create a portable backup archive", Long: "Create a portable backup archive. Stop Dottie first so both embedded databases have a consistent checkpoint.", RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		path, err := backup.Create(cfg, output)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	}}
	command.Flags().StringVarP(&output, "output", "o", "", "backup archive path")
	return command
}

func restoreCommand(load configLoader) *cobra.Command {
	var force bool
	command := &cobra.Command{Use: "restore BACKUP", Args: cobra.ExactArgs(1), Short: "Restore a Dottie backup", RunE: func(_ *cobra.Command, args []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		return backup.Restore(cfg, args[0], force)
	}}
	command.Flags().BoolVar(&force, "force", false, "replace existing data")
	return command
}

func adminCommand(load configLoader) *cobra.Command {
	var email string
	admin := &cobra.Command{Use: "admin", Short: "Administrative recovery commands"}
	reset := &cobra.Command{Use: "reset-password", Short: "Reset the local administrator password", RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := load()
		if err != nil {
			return err
		}
		if email == "" {
			fmt.Print("Email: ")
			value, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			email = strings.TrimSpace(value)
		}
		password := os.Getenv("DOTTIE_NEW_PASSWORD")
		if password == "" {
			fmt.Print("New password: ")
			value, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password = string(value)
		}
		ctx := context.Background()
		client, err := sqlitestore.Open(ctx, cfg.SQLitePath())
		if err != nil {
			return err
		}
		defer client.Close()
		if err := domain.NewService(client.Queries).ResetPassword(ctx, email, password); err != nil {
			return err
		}
		fmt.Println("Password reset successfully.")
		return nil
	}}
	reset.Flags().StringVar(&email, "email", "", "administrator email address")
	admin.AddCommand(reset)
	return admin
}

func openAPICommand() *cobra.Command {
	var output string
	command := &cobra.Command{Use: "openapi", Hidden: true, RunE: func(_ *cobra.Command, _ []string) error {
		api := httpapi.New(httpapi.Dependencies{Logger: slog.Default(), Version: version})
		data, err := api.API.OpenAPI().DowngradeYAML()
		if err != nil {
			return err
		}
		if output != "" {
			if err := os.WriteFile(output, data, 0o644); err != nil {
				return fmt.Errorf("write OpenAPI document: %w", err)
			}
			return nil
		}
		_, err = os.Stdout.Write(data)
		return err
	}}
	command.Flags().StringVarP(&output, "output", "o", "", "write the OpenAPI document to a file")
	return command
}

func newLogger(level string) *slog.Logger {
	logLevel := slog.LevelInfo
	if level == "debug" {
		logLevel = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
}

func checkResult(name string, err error) map[string]any {
	result := map[string]any{"name": name, "ok": err == nil}
	if err != nil {
		result["error"] = err.Error()
	}
	return result
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
