package main

import (
	"database/sql"
	"testing"

	"dataray/internal/model"
)

// TestBuildSQLBasicBarChart tests building SQL for basic bar chart
func TestBuildSQLBasicBarChart(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "sales", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis":     "region",
		"yAxis":     "amount",
		"chartType": "bar",
	}

	sql, err := buildSQL(ds, config, "postgresql")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT region, amount FROM sales GROUP BY region ORDER BY region LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}

// TestBuildSQLPieChart tests building SQL for pie chart
func TestBuildSQLPieChart(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "products", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis":     "category",
		"yAxis":     "count",
		"chartType": "pie",
	}

	sql, err := buildSQL(ds, config, "postgresql")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT category, SUM(count) as value FROM products GROUP BY category ORDER BY category LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}

// TestBuildSQLLineChart tests building SQL for line chart
func TestBuildSQLLineChart(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "orders", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis":     "date",
		"yAxis":     "quantity",
		"chartType": "line",
	}

	sql, err := buildSQL(ds, config, "postgresql")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT date, quantity FROM orders ORDER BY date LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}

// TestBuildSQLWithSQLQuery tests building SQL with custom SQL query
func TestBuildSQLWithSQLQuery(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "sql",
		QuerySQL:  sql.NullString{String: "SELECT region, SUM(amount) as amount FROM sales GROUP BY region", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis":     "region",
		"yAxis":     "amount",
		"chartType": "bar",
	}

	sql, err := buildSQL(ds, config, "postgresql")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT region, amount FROM (SELECT region, SUM(amount) as amount FROM sales GROUP BY region) AS subq GROUP BY region ORDER BY region LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}

// TestBuildSQLErrorMissingXAxis tests error when xAxis is missing
func TestBuildSQLErrorMissingXAxis(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "sales", Valid: true},
	}
	config := map[string]interface{}{
		"yAxis":     "amount",
		"chartType": "bar",
	}

	_, err := buildSQL(ds, config, "postgresql")
	if err == nil {
		t.Error("expected error when xAxis is missing")
	}
}

// TestBuildSQLErrorNoTableOrQuery tests error when no table or query defined
func TestBuildSQLErrorNoTableOrQuery(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
	}
	config := map[string]interface{}{
		"xAxis":     "region",
		"yAxis":     "amount",
		"chartType": "bar",
	}

	_, err := buildSQL(ds, config, "postgresql")
	if err == nil {
		t.Error("expected error when no table or query defined")
	}
}

// TestBuildSQLEmptyConfig tests buildSQL with empty config
func TestBuildSQLEmptyConfig(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "sales", Valid: true},
	}
	config := map[string]interface{}{}

	_, err := buildSQL(ds, config, "postgresql")
	if err == nil {
		t.Error("expected error when xAxis is missing in empty config")
	}
}

// TestBuildSQLOnlyXAxis tests buildSQL with only xAxis
func TestBuildSQLOnlyXAxis(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "sales", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis": "region",
	}

	sql, err := buildSQL(ds, config, "postgresql")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT region FROM sales ORDER BY region LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}

// TestBuildSQLClickHouse tests building SQL for ClickHouse
func TestBuildSQLClickHouse(t *testing.T) {
	ds := &model.Dataset{
		QueryType: "table",
		TableName: sql.NullString{String: "sales", Valid: true},
	}
	config := map[string]interface{}{
		"xAxis":     "region",
		"yAxis":     "amount",
		"chartType": "bar",
	}

	sql, err := buildSQL(ds, config, "clickhouse")
	if err != nil {
		t.Fatalf("buildSQL failed: %v", err)
	}

	expected := "SELECT region, amount FROM sales GROUP BY region ORDER BY region LIMIT 1000"
	if sql != expected {
		t.Errorf("expected SQL: %s, got: %s", expected, sql)
	}
}
