package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataray/internal/response"

	"github.com/gin-gonic/gin"
)

// TestHealthRouteUsesUnifiedResponse verifies the operational health endpoint also goes through
// the shared response envelope instead of bypassing response normalization with direct c.JSON.
func TestHealthRouteUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerHealthRoute(router)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}
	if payload["code"] != float64(response.CodeSuccess) {
		t.Fatalf("expected code %d, got %v", response.CodeSuccess, payload["code"])
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", payload["data"])
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", data["status"])
	}
}
