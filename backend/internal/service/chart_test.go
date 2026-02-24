package service

import (
	"database/sql"
	"dataray/internal/model"
	"testing"
)

func TestInferRole(t *testing.T) {
	tests := []struct {
		dataType string
		expected string
	}{
		{"int", "metric"},
		{"integer", "metric"},
		{"float", "metric"},
		{"double", "metric"},
		{"decimal", "metric"},
		{"numeric", "metric"},
		{"real", "metric"},
		{"number", "metric"},
		{"varchar", "dimension"},
		{"string", "dimension"},
		{"date", "dimension"},
		{"timestamp", "dimension"},
		{"text", "dimension"},
		{"", "dimension"},
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			result := inferRole(tt.dataType)
			if result != tt.expected {
				t.Errorf("inferRole(%q) = %q, want %q", tt.dataType, result, tt.expected)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	token1 := generateToken()
	token2 := generateToken()

	if token1 == "" {
		t.Error("generateToken() returned empty token")
	}
	if token1 == token2 {
		t.Error("generateToken() returned same token twice")
	}
	if len(token1) < 10 {
		t.Errorf("generateToken() token too short: %d", len(token1))
	}
}

func TestParseDatasetColumns(t *testing.T) {
	tests := []struct {
		name        string
		columnsJSON string
		wantErr     bool
	}{
		{"empty", "", false},
		{"empty array", "[]", false},
		{"valid", `[{"name": "id", "type": "int"}]`, false},
		{"invalid json", "{invalid}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDatasetColumns(tt.columnsJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDatasetColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveExpr(t *testing.T) {
	columns := []model.DatasetColumn{
		{Name: "id", Expr: "`id`"},
		{Name: "name", Expr: "`name`"},
	}

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{"empty", "", ""},
		{"no placeholder", "SELECT *", "SELECT *"},
		{"single placeholder", "[id]", "`id`"},
		{"multiple placeholders", "[id], [name]", "`id`, `name`"},
		{"no match", "[other]", "[other]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveExpr(tt.expr, columns)
			if err != nil {
				t.Errorf("resolveExpr() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("resolveExpr() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildSQL_MissingXAxis(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "test_table", Valid: true},
	}
	config := map[string]interface{}{}

	_, err := buildSQL(ds, config)
	if err == nil {
		t.Error("buildSQL() expected error for missing xAxis")
	}
}
