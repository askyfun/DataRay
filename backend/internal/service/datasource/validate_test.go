package datasource

import (
	"testing"
)

func TestIsValidSQLIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid identifiers
		{"simple table name", "users", true},
		{"table with underscore", "user_accounts", true},
		{"table with numbers", "table123", true},
		{"schema qualified", "public.users", true},
		{"schema and table with underscores", "my_schema.my_table", true},
		{"mixed case", "UserTable", true},
		{"single char", "a", true},
		{"number only", "123", true},

		// Invalid identifiers
		{"empty string", "", false},
		{"space in name", "user accounts", false},
		{"single quote", "user'name", false},
		{"semicolon", "users;", false},
		{"SQL injection with DROP", "users; DROP TABLE users", false},
		{"parentheses", "users()", false},
		{"asterisk", "users*", false},
		{"hyphen", "user-accounts", false},
		{"backtick", "users`", false},
		{"double quote", `users"`, false},
		{"backslash", `users\`, false},
		{"newline", "users\n", false},
		{"tab", "users\t", false},
		{"unicode", "用户", false},
		{"percent sign", "users%", false},
		{"equals sign", "users=1", false},
		{"ampersand", "users&roles", false},
		{"pipe", "users|roles", false},
		{"exclamation", "users!", false},
		{"at sign", "users@db", false},
		{"hash", "users#", false},
		{"dollar sign", "users$", false},
		{"caret", "users^", false},
		{"curly braces", "users{}", false},
		{"square brackets", "users[]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSQLIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("isValidSQLIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSortOrder(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty defaults to ASC", "", "ASC"},
		{"uppercase ASC", "ASC", "ASC"},
		{"uppercase DESC", "DESC", "DESC"},
		{"lowercase asc", "asc", "ASC"},
		{"lowercase desc", "desc", "DESC"},
		{"mixed case Desc", "Desc", "DESC"},
		{"mixed case Asc", "Asc", "ASC"},
		{"invalid value defaults to ASC", "INVALID", "ASC"},
		{"random string defaults to ASC", "xyz", "ASC"},
		{"numeric string defaults to ASC", "123", "ASC"},
		{"partial match defaults to ASC", "DES", "ASC"},
		{"extra chars defaults to ASC", "DESCC", "ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSortOrder(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSortOrder(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		val   string
		want  bool
	}{
		{"found in slice", []string{"id", "name"}, "id", true},
		{"not found", []string{"id", "name"}, "email", false},
		{"empty slice", []string{}, "id", false},
		{"nil slice", nil, "id", false},
		{"empty val found", []string{""}, "", true},
		{"case sensitive", []string{"ID"}, "id", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.slice, tt.val)
			if got != tt.want {
				t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.val, got, tt.want)
			}
		})
	}
}
