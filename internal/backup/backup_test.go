package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/d3cie/dottie/internal/config"
)

func TestBackupRoundTrip(t *testing.T) {
	source := t.TempDir()
	cfg := config.Config{DataDir: source, Address: "127.0.0.1:8080", LogLevel: "info"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"dottie.sqlite": "sqlite-data", "analytics.duckdb": "duck-data"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := Create(cfg, archive); err != nil {
		t.Fatal(err)
	}

	destination := t.TempDir()
	if err := Restore(config.Config{DataDir: destination}, archive, false); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"dottie.sqlite": "sqlite-data", "analytics.duckdb": "duck-data"} {
		got, err := os.ReadFile(filepath.Join(destination, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s: got %q, want %q", name, got, want)
		}
	}
}
