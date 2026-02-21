package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file - using yaml tags from embedded RestConf
	content := `
Name: testapp
Host: 127.0.0.1
Port: 9090
Database:
  Url: postgres://user:pass@localhost:5432/testdb?sslmode=disable
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test LoadConfig
	var cfg Config
	if err := cfg.LoadConfig(tmpFile.Name()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if cfg.Name != "testapp" {
		t.Errorf("expected Name=testapp, got %s", cfg.Name)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host=127.0.0.1, got %s", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", cfg.Port)
	}
	if cfg.Database.Url != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("expected Database.Url, got %s", cfg.Database.Url)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	var cfg Config
	err := cfg.LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	content := `
Name: testapp
Host: 127.0.0.1
Port: "not-a-number"
Database:
  Url: postgres://user:pass@localhost:5432/testdb
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	var cfg Config
	err = cfg.LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	// Create an empty temporary file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	var cfg Config
	err = cfg.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("empty config should be valid: %v", err)
	}
}
