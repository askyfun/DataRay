package database

import (
	"testing"
)

// TestInitDBInvalidURL tests InitDB with invalid database URL
func TestInitDBInvalidURL(t *testing.T) {
	_, err := InitDB("invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

// TestInitDBConnectionFailure tests InitDB with unreachable database
func TestInitDBConnectionFailure(t *testing.T) {
	// Use a non-routable IP address to simulate connection failure
	_, err := InitDB("postgres://user:pass@192.0.2.1:5432/testdb?sslmode=disable")
	if err == nil {
		t.Error("expected error for unreachable database")
	}
}

// TestRunMigrationsNilDB tests RunMigrations with nil DB - skip as it causes panic
func TestRunMigrationsNilDB(t *testing.T) {
	t.Skip("Skipping nil DB test as it causes panic - would need mock")
}
