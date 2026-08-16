package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	defaultAddress = "127.0.0.1:8080"
	fileName       = "config.json"
)

type Config struct {
	DataDir  string `json:"data_dir"`
	Address  string `json:"address"`
	BaseURL  string `json:"base_url,omitempty"`
	LogLevel string `json:"log_level"`
}

func Load(path string) (Config, error) {
	dataDir, err := defaultDataDir()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{DataDir: dataDir, Address: defaultAddress, LogLevel: "info"}
	if path == "" {
		path = filepath.Join(dataDir, fileName)
	}

	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode config %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	applyEnvironment(&cfg)
	if cfg.DataDir == "" {
		return Config{}, errors.New("data directory cannot be empty")
	}
	if cfg.Address == "" {
		return Config{}, errors.New("listen address cannot be empty")
	}
	return cfg, nil
}

func Path(cfg Config) string {
	return filepath.Join(cfg.DataDir, fileName)
}

func Save(cfg Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(Path(cfg), data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c Config) SQLitePath() string {
	return filepath.Join(c.DataDir, "dottie.sqlite")
}

func (c Config) DuckDBPath() string {
	return filepath.Join(c.DataDir, "analytics.duckdb")
}

func (c Config) PublicURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	address := c.Address
	if strings.HasPrefix(address, "0.0.0.0:") {
		address = "127.0.0.1:" + strings.TrimPrefix(address, "0.0.0.0:")
	} else if strings.HasPrefix(address, "[::]:") {
		address = "127.0.0.1:" + strings.TrimPrefix(address, "[::]:")
	}
	return "http://" + address
}

func applyEnvironment(cfg *Config) {
	if value := os.Getenv("DOTTIE_DATA_DIR"); value != "" {
		cfg.DataDir = value
	}
	if value := os.Getenv("DOTTIE_ADDRESS"); value != "" {
		cfg.Address = value
	}
	if value := os.Getenv("DOTTIE_BASE_URL"); value != "" {
		cfg.BaseURL = value
	}
	if value := os.Getenv("DOTTIE_LOG_LEVEL"); value != "" {
		cfg.LogLevel = value
	}
	if value := os.Getenv("PORT"); value != "" && os.Getenv("DOTTIE_ADDRESS") == "" {
		if _, err := strconv.Atoi(value); err == nil {
			cfg.Address = "0.0.0.0:" + value
		}
	}
}

func defaultDataDir() (string, error) {
	if value := os.Getenv("DOTTIE_DATA_DIR"); value != "" {
		return value, nil
	}

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "dottie"), nil
	case "windows":
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve config directory: %w", err)
		}
		return filepath.Join(base, "dottie"), nil
	default:
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve config directory: %w", err)
		}
		return filepath.Join(base, "dottie"), nil
	}
}
