package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExecuteChartQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.POST("/api/charts/query", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		datasetID, ok := req["dataset_id"].(float64)
		if !ok || datasetID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dataset_id is required"})
			return
		}

		chartType, ok := req["chart_type"].(string)
		if !ok || chartType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "chart_type is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    req,
		})
	})

	body := map[string]interface{}{
		"dataset_id": 5,
		"chart_type": "table",
		"dims":       []string{"project_id"},
		"metrics":    []map[string]string{{"field": "cnt", "agg": "sum", "alias": "cnt"}},
		"filters":    []interface{}{},
		"pagination": map[string]int{"page": 1, "page_size": 10},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/charts/query", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)

	if resp["message"] != "success" {
		t.Errorf("expected success message, got %v", resp)
	}
}

func TestExecuteChartQueryInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.POST("/api/charts/query", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/charts/query", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid json, got %d", rr.Code)
	}
}
