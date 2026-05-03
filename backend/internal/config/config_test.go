package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
Name = "testapp"
Host = "127.0.0.1"
Port = 9090

[Database]
Url = "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
`
	tmpFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg := &Config{}
	if err := cfg.LoadConfig(tmpFile.Name()); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "testapp" {
		t.Errorf("expected name 'testapp', got '%s'", cfg.Name)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Database.Url != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("expected database url, got '%s'", cfg.Database.Url)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	cfg := &Config{}
	err := cfg.LoadConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for nonexistent config file")
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	content := `
Name = "testapp"
Host = "127.0.0.1"
Port = 9090

[Database]
Url = "invalid url"
`
	tmpFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg := &Config{}
	if err := cfg.LoadConfig(tmpFile.Name()); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
}
