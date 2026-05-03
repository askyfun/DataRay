package query

import (
	"encoding/json"
	"testing"
)

func TestBunQueryBuilder_BasicQuery(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"test_table",
		SourceTypeTable,
		[]string{"project_id"},
		[]MetricConfig{{Field: "amount", Agg: AggSum, Alias: "total_amount"}},
		[]FilterConfig{},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}

	expected := "SELECT project_id, SUM(amount) AS total_amount FROM test_table GROUP BY project_id"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBunQueryBuilder_WithPagination(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status"},
		[]MetricConfig{{Field: "id", Agg: AggCount, Alias: "count"}},
		[]FilterConfig{},
		nil,
		&Pagination{Page: 2, PageSize: 20},
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT status, COUNT(id) AS count FROM orders GROUP BY status LIMIT 20 OFFSET 20"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	// LIMIT/OFFSET uses direct embedding for ClickHouse compatibility (not parameterized)
}

func TestBunQueryBuilder_WithFilters(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"users",
		SourceTypeTable,
		[]string{"city"},
		[]MetricConfig{{Field: "age", Agg: AggAvg, Alias: "avg_age"}},
		[]FilterConfig{
			{Field: "status", Op: FilterEq, Value: "active"},
			{Field: "age", Op: FilterGt, Value: 18},
		},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT city, AVG(age) AS avg_age FROM users WHERE status = ? AND age > ? GROUP BY city"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestBunQueryBuilder_WithSort(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"sales",
		SourceTypeTable,
		[]string{"product"},
		[]MetricConfig{{Field: "revenue", Agg: AggSum, Alias: "total_revenue"}},
		[]FilterConfig{},
		&SortConfig{Field: "total_revenue", Order: "desc"},
		nil,
	)

	sql, _ := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)

	expected := "SELECT product, SUM(revenue) AS total_revenue FROM sales GROUP BY product ORDER BY total_revenue desc"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBunQueryBuilder_CountQuery(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status"},
		[]MetricConfig{{Field: "id", Agg: AggCount, Alias: "count"}},
		[]FilterConfig{},
		nil,
		nil,
	)

	sql, args := qb.BuildCountQuery(ast)

	t.Logf("Generated Count SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT COUNT(*) AS _total FROM (SELECT 1 FROM orders GROUP BY status) AS _count_query"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBunQueryBuilder_WithSQLSource(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"SELECT * FROM orders WHERE created_at > '2024-01-01'",
		SourceTypeSQL,
		[]string{"status"},
		[]MetricConfig{{Field: "amount", Agg: AggSum, Alias: "total"}},
		[]FilterConfig{},
		nil,
		nil,
	)

	sql, _ := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)

	expected := "SELECT status, SUM(amount) AS total FROM (SELECT * FROM orders WHERE created_at > '2024-01-01') AS _subq GROUP BY status"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBunSQLBuilder_BuildSelect(t *testing.T) {
	qb := NewBunQueryBuilder()
	qb.columnMappings = map[string]string{
		"cnt": "count(*)",
	}

	ast := qb.Build(
		"test_table",
		SourceTypeTable,
		[]string{"project_id"},
		[]MetricConfig{{Field: "cnt", Agg: AggSum, Alias: "cnt"}},
		[]FilterConfig{},
		nil,
		&Pagination{Page: 1, PageSize: 10},
	)

	builder := NewBunSQLBuilder(DialectPostgreSQL)
	sql := builder.BuildSelect(ast)

	t.Logf("Generated SQL: %s", sql)

	expected := "SELECT project_id, count(*) AS cnt FROM test_table GROUP BY project_id LIMIT 10 OFFSET 0"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBunQueryBuilder_SafeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_field", "normal_field"},
		{"field_name", "field_name"},
		{"field; DROP TABLE", "_invalid_identifier"},
		{"field' OR '1'='1", "_invalid_identifier"},
		{"field--comment", "_invalid_identifier"},
		{"  spaces  ", "spaces"},
	}

	for _, tt := range tests {
		result := safeIdentifier(tt.input)
		if result != tt.expected {
			t.Errorf("safeIdentifier(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestBunQueryBuilder_ColumnMappings(t *testing.T) {
	qb := NewBunQueryBuilder()

	columns := `[{"name":"revenue","expr":"SUM(amount)","role":"metric","type":"bigint"}]`
	if err := qb.WithColumnMappings(columns); err != nil {
		t.Fatalf("Failed to parse column mappings: %v", err)
	}

	ast := qb.Build(
		"sales",
		SourceTypeTable,
		[]string{"product"},
		[]MetricConfig{{Field: "revenue", Agg: AggSum}},
		[]FilterConfig{},
		nil,
		nil,
	)

	sql, _ := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)

	expected := "SELECT product, SUM(revenue) AS revenue FROM sales GROUP BY product"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestBuildQueryStringWithBun(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"test_table",
		SourceTypeTable,
		[]string{"category"},
		[]MetricConfig{{Field: "price", Agg: AggAvg, Alias: "avg_price"}},
		[]FilterConfig{},
		nil,
		nil,
	)

	selectSQL, countSQL := BuildQueryStringWithBun(DialectPostgreSQL, ast)

	t.Logf("Select SQL: %s", selectSQL)
	t.Logf("Count SQL: %s", countSQL)

	expectedSelect := "SELECT category, AVG(price) AS avg_price FROM test_table GROUP BY category"
	if selectSQL != expectedSelect {
		t.Errorf("Expected select:\n%s\nGot:\n%s", expectedSelect, selectSQL)
	}

	expectedCount := "SELECT COUNT(*) AS _total FROM (SELECT 1 FROM test_table GROUP BY category) AS _count_query"
	if countSQL != expectedCount {
		t.Errorf("Expected count:\n%s\nGot:\n%s", expectedCount, countSQL)
	}
}

func TestBunQueryBuilder_WithFilterIn(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"products",
		SourceTypeTable,
		[]string{"category"},
		[]MetricConfig{{Field: "id", Agg: AggCount, Alias: "count"}},
		[]FilterConfig{
			{Field: "status", Op: FilterIn, Value: []any{"active", "pending", "draft"}},
		},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT category, COUNT(id) AS count FROM products WHERE status IN (?, ?, ?) GROUP BY category"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestBunQueryBuilder_WithFilterBetween(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status"},
		[]MetricConfig{{Field: "amount", Agg: AggSum, Alias: "total"}},
		[]FilterConfig{
			{Field: "created_at", Op: FilterBetween, Value: "2024-01-01", ValueEnd: "2024-12-31"},
		},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT status, SUM(amount) AS total FROM orders WHERE created_at BETWEEN ? AND ? GROUP BY status"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestBunQueryBuilder_WithFilterLike(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"users",
		SourceTypeTable,
		[]string{"city"},
		[]MetricConfig{{Field: "id", Agg: AggCount, Alias: "count"}},
		[]FilterConfig{
			{Field: "name", Op: FilterLike, Value: "John"},
		},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT city, COUNT(id) AS count FROM users WHERE name LIKE ? GROUP BY city"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestBunQueryBuilder_WithFilterNull(t *testing.T) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status"},
		[]MetricConfig{{Field: "id", Agg: AggCount, Alias: "count"}},
		[]FilterConfig{
			{Field: "deleted_at", Op: FilterIsNull},
		},
		nil,
		nil,
	)

	sql, args := qb.BuildSelectQuery(ast)

	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", args)

	expected := "SELECT status, COUNT(id) AS count FROM orders WHERE deleted_at IS NULL GROUP BY status"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func BenchmarkBunQueryBuilder_BuildSelect(b *testing.B) {
	qb := NewBunQueryBuilder()

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status", "category"},
		[]MetricConfig{
			{Field: "amount", Agg: AggSum, Alias: "total_amount"},
			{Field: "quantity", Agg: AggAvg, Alias: "avg_quantity"},
		},
		[]FilterConfig{
			{Field: "status", Op: FilterEq, Value: "active"},
			{Field: "amount", Op: FilterGt, Value: 100},
		},
		&SortConfig{Field: "total_amount", Order: "desc"},
		&Pagination{Page: 1, PageSize: 20},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qb.BuildSelectQuery(ast)
	}
}

func BenchmarkOldBuilder_BuildSelect(b *testing.B) {
	qb := NewQueryBuilder()
	qb.WithColumnMappings(`[{"name":"revenue","expr":"SUM(amount)","role":"metric","type":"bigint"}]`)

	ast := qb.Build(
		"orders",
		SourceTypeTable,
		[]string{"status", "category"},
		[]MetricConfig{
			{Field: "amount", Agg: AggSum, Alias: "total_amount"},
			{Field: "quantity", Agg: AggAvg, Alias: "avg_quantity"},
		},
		[]FilterConfig{
			{Field: "status", Op: FilterEq, Value: "active"},
			{Field: "amount", Op: FilterGt, Value: 100},
		},
		&SortConfig{Field: "total_amount", Order: "desc"},
		&Pagination{Page: 1, PageSize: 20},
	)

	builder := NewPostgreSQLBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.BuildSelect(ast)
	}
}

func init() {
	_ = json.Marshal
}
