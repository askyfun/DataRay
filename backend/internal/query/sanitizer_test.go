package query

import (
	"testing"
)

func TestEscapeValue_String(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"normal string", "hello", "hello"},
		{"string with single quote", "it's", "it''s"},
		{"string with multiple quotes", "it's a 'test'", "it''s a ''test''"},
		{"empty string", "", ""},
		{"string with sql injection attempt", "'; DROP TABLE users; --", "''; DROP TABLE users; --"},
		{"string with LIKE wildcards", "test%value", "test%value"},
		{"int value", 42, "42"},
		{"int64 value", int64(123456789012345), "123456789012345"},
		{"float value", 3.14, "3.14"},
		{"float precision", 1.23456789012345, "1.23456789012345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeValue(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"it's", "it''s"},
		{"it''s", "it''''s"},
		{"''", "''''"},
		{"", ""},
		{"'; DELETE FROM users; --", "''; DELETE FROM users; --"},
		{"'; 1=1 --", "''; 1=1 --"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeValue_NumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"int", int(42), "42"},
		{"int8", int8(8), "8"},
		{"int16", int16(16), "16"},
		{"int32", int32(32), "32"},
		{"int64", int64(64), "64"},
		{"uint", uint(100), "100"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"float32", float32(3.14), "3.14"},
		{"float64", float64(2.71828), "2.71828"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeValue(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeValue(%T(%v)) = %q, want %q", tt.input, tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeValue_Nil(t *testing.T) {
	result := EscapeValue(nil)
	if result != "" {
		t.Errorf("EscapeValue(nil) = %q, want %q", result, "")
	}
}

func TestEscapeValue_UnknownType(t *testing.T) {
	type CustomStruct struct {
		Field string
	}
	input := CustomStruct{Field: "test"}
	result := EscapeValue(input)
	if result == "" {
		t.Errorf("EscapeValue(CustomStruct) = %q, want non-empty string", result)
	}
}
