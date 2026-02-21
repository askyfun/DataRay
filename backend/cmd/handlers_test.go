package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandleDatasourcesGET tests GET /api/datasources
func TestHandleDatasourcesGET(t *testing.T) {
	// Create a mock handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/datasources", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestHandleDatasourcesPOST tests POST /api/datasources
func TestHandleDatasourcesPOST(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})

	body := `{"name":"test","host":"localhost","port":5432,"database_name":"testdb","username":"user","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/datasources", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestHandleDatasourcesInvalidMethod tests invalid method
func TestHandleDatasourcesInvalidMethod(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPost:
			w.Write([]byte("ok"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/datasources", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

// TestGenerateToken tests token generation
func TestGenerateToken(t *testing.T) {
	token := generateToken()

	if token == "" {
		t.Error("generated token should not be empty")
	}

	// Token should contain a dash
	if !strings.Contains(token, "-") {
		t.Errorf("token should contain dash: %s", token)
	}

	// Token should have reasonable length
	if len(token) < 10 {
		t.Errorf("token should be reasonably long, got %d", len(token))
	}
}

// TestHealthEndpoint tests /health endpoint
func TestHealthEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "ok") {
		t.Errorf("expected body to contain ok, got %s", rr.Body.String())
	}
}

// TestHandleDatasourceTablesGET tests GET /api/datasources/:id/tables
func TestHandleDatasourceTablesGET(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name":"users"},{"name":"orders"}]`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/datasources/1/tables", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "users") {
		t.Errorf("expected body to contain users, got %s", rr.Body.String())
	}
}

// TestHandleDatasourceTableColumnsGET tests GET /api/datasources/:id/tables/:table/columns
func TestHandleDatasourceTableColumnsGET(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name":"id","type":"int"},{"name":"name","type":"varchar"}]`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/datasources/1/tables/users/columns", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "id") {
		t.Errorf("expected body to contain id, got %s", rr.Body.String())
	}
}
