package query

import (
	"testing"
)

func TestMySQLBuilder_WithMetrics(t *testing.T) {
	qb := NewQueryBuilder()
	qb.WithColumnMappings(`[{"name":"cnt","expr":"count(*)","role":"metric","type":"bigint"}]`)

	ast := qb.Build(
		"test_table",
		SourceTypeTable,
		[]string{"project_id"},
		[]MetricConfig{{Field: "cnt", Agg: AggSum, Alias: "cnt"}},
		[]FilterConfig{},
		nil,
		&Pagination{Page: 1, PageSize: 10},
	)

	builder := NewMySQLBuilder()
	sql := builder.BuildSelect(ast)

	t.Logf("Generated SQL: %s", sql)

	expected := "SELECT project_id, count(*) AS cnt FROM test_table GROUP BY project_id LIMIT 10 OFFSET 0"
	if sql != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, sql)
	}
}

func TestMySQLBuilder_WithoutDims(t *testing.T) {
	qb := NewQueryBuilder()
	qb.WithColumnMappings(`[{"name":"cnt","expr":"count(*)","role":"metric","type":"bigint"}]`)

	ast := qb.Build(
		"test_table",
		SourceTypeTable,
		[]string{},
		[]MetricConfig{{Field: "cnt", Agg: AggSum, Alias: "cnt"}},
		[]FilterConfig{},
		nil,
		nil,
	)

	builder := NewMySQLBuilder()
	sql := builder.BuildSelect(ast)

	t.Logf("Generated SQL: %s", sql)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
}
