package backup

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/d3cie/dottie/internal/config"
)

type manifest struct {
	Format    int       `json:"format"`
	CreatedAt time.Time `json:"created_at"`
}

func Create(cfg config.Config, destination string) (string, error) {
	if destination == "" {
		destination = fmt.Sprintf("dottie-backup-%s.tar.gz", time.Now().UTC().Format("20060102-150405"))
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	remove := true
	defer func() {
		_ = file.Close()
		if remove {
			_ = os.Remove(destination)
		}
	}()
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	manifestData, _ := json.Marshal(manifest{Format: 1, CreatedAt: time.Now().UTC()})
	if err := writeBytes(tarWriter, "manifest.json", manifestData, 0o600); err != nil {
		return "", err
	}
	for _, name := range []string{"config.json", "dottie.sqlite", "analytics.duckdb"} {
		if err := writeFile(tarWriter, filepath.Join(cfg.DataDir, name), name); err != nil {
			return "", err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("finalize backup archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return "", fmt.Errorf("compress backup archive: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close backup archive: %w", err)
	}
	remove = false
	absolute, _ := filepath.Abs(destination)
	return absolute, nil
}

func Restore(cfg config.Config, source string, force bool) error {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("read backup compression: %w", err)
	}
	defer gzipReader.Close()

	if !force {
		for _, name := range []string{"dottie.sqlite", "analytics.duckdb"} {
			if _, err := os.Stat(filepath.Join(cfg.DataDir, name)); err == nil {
				return errors.New("data already exists; pass --force to replace it")
			}
		}
	}
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	tmp, err := os.MkdirTemp(cfg.DataDir, ".restore-*")
	if err != nil {
		return fmt.Errorf("create restore staging directory: %w", err)
	}
	defer os.RemoveAll(tmp)

	reader := tar.NewReader(gzipReader)
	seenManifest := false
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read backup entry: %w", err)
		}
		name := filepath.Base(header.Name)
		if name != header.Name || !allowedFile(name) {
			return fmt.Errorf("backup contains unexpected file %q", header.Name)
		}
		data, err := io.ReadAll(io.LimitReader(reader, 100<<30))
		if err != nil {
			return fmt.Errorf("read backup file %s: %w", name, err)
		}
		if name == "manifest.json" {
			var value manifest
			if err := json.Unmarshal(data, &value); err != nil || value.Format != 1 {
				return errors.New("unsupported backup format")
			}
			seenManifest = true
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0o600); err != nil {
			return fmt.Errorf("stage backup file %s: %w", name, err)
		}
	}
	if !seenManifest {
		return errors.New("backup manifest is missing")
	}
	for _, name := range []string{"config.json", "dottie.sqlite", "analytics.duckdb"} {
		staged := filepath.Join(tmp, name)
		if _, err := os.Stat(staged); err != nil {
			return fmt.Errorf("backup is missing %s", name)
		}
		if err := os.Rename(staged, filepath.Join(cfg.DataDir, name)); err != nil {
			return fmt.Errorf("restore %s: %w", name, err)
		}
	}
	return nil
}

func allowedFile(name string) bool {
	return name == "manifest.json" || name == "config.json" || name == "dottie.sqlite" || name == "analytics.duckdb"
}

func writeFile(writer *tar.Writer, source, name string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", name, err)
	}
	if strings.HasSuffix(name, ".wal") {
		return nil
	}
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()
	if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: info.Size(), ModTime: info.ModTime()}); err != nil {
		return fmt.Errorf("write %s header: %w", name, err)
	}
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func writeBytes(writer *tar.Writer, name string, data []byte, mode int64) error {
	if err := writer.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(data)), ModTime: time.Now()}); err != nil {
		return fmt.Errorf("write %s header: %w", name, err)
	}
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}
