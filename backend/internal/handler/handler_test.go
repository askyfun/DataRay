package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "users", true},
		{"valid with underscore", "my_table", true},
		{"valid with numbers", "table1", true},
		{"valid mixed", "my_table_123", true},
		{"invalid empty", "", false},
		{"invalid starts with number", "1table", false},
		{"invalid special chars", "my-table", false},
		{"invalid space", "my table", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseID_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "123"}}

	id, err := parseID(c, "id")
	if err != nil {
		t.Errorf("parseID() error = %v", err)
	}
	if id != 123 {
		t.Errorf("parseID() = %d, want 123", id)
	}
}

func TestParseID_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	_, err := parseID(c, "id")
	if err == nil {
		t.Error("parseID() expected error for invalid id")
	}
}

func TestParseID_Negative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "-1"}}

	_, err := parseID(c, "id")
	if err == nil {
		t.Error("parseID() expected error for negative id")
	}
}

func TestParseID_Zero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "0"}}

	_, err := parseID(c, "id")
	if err == nil {
		t.Error("parseID() expected error for zero id")
	}
}

func TestGetPaginationParams_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	limit, offset := getPaginationParams(c)
	if limit != 100 {
		t.Errorf("default limit = %d, want 100", limit)
	}
	if offset != 0 {
		t.Errorf("default offset = %d, want 0", offset)
	}
}

func TestGetPaginationParams_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/?limit=10&offset=20", nil)

	limit, offset := getPaginationParams(c)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 20 {
		t.Errorf("offset = %d, want 20", offset)
	}
}

func TestGetPaginationParams_LimitOverflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/?limit=2000", nil)

	limit, _ := getPaginationParams(c)
	if limit != 100 {
		t.Errorf("limit should be capped at 100, got %d", limit)
	}
}
