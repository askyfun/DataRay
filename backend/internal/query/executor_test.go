package query

import (
	"context"
	"database/sql"
	"testing"

	"dataray/internal/datasource"
	"dataray/internal/model"
)

type MockConnection struct {
	rows     []map[string]any
	queryErr error
}

func (m *MockConnection) Execute(ctx context.Context, sql string) (*datasource.QueryResult, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return &datasource.QueryResult{Rows: m.rows}, nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) Ping(ctx context.Context) error {
	return nil
}

func (m *MockConnection) GetTables(ctx context.Context) ([]datasource.TableInfo, error) {
	return nil, nil
}

func (m *MockConnection) GetColumns(ctx context.Context, tableName string) ([]datasource.ColumnInfo, error) {
	return nil, nil
}

func TestExecutor_Execute_EmptyBaseQuery(t *testing.T) {
	dataset := &model.Dataset{
		ID:        1,
		Name:      "Test Dataset",
		QueryType: "table",
		TableName: sql.NullString{Valid: false},
		QuerySQL:  sql.NullString{Valid: false},
	}

	ds := &model.Datasource{
		ID:   1,
		Name: "Test DS",
		Type: "postgresql",
	}

	conn := &MockConnection{}
	executor := NewExecutor(conn, dataset, ds)

	req := &ChartQueryRequest{
		DatasetID: 1,
		ChartType: ChartTypeBar,
		Dims:      []string{"category"},
		Metrics:   []MetricConfig{{Field: "value", Agg: AggSum}},
	}

	_, err := executor.Execute(context.Background(), req)

	if err == nil {
		t.Error("Expected error for empty base query, got nil")
	}
}

func TestExecutor_Execute_ValidBaseQuery(t *testing.T) {
	dataset := &model.Dataset{
		ID:        1,
		Name:      "Test Dataset",
		QueryType: "table",
		TableName: sql.NullString{String: "test_table", Valid: true},
		QuerySQL:  sql.NullString{Valid: false},
	}

	ds := &model.Datasource{
		ID:   1,
		Name: "Test DS",
		Type: "postgresql",
	}

	conn := &MockConnection{
		rows: []map[string]any{
			{"category": "A", "value": 100},
			{"category": "B", "value": 200},
		},
	}
	executor := NewExecutor(conn, dataset, ds)

	req := &ChartQueryRequest{
		DatasetID: 1,
		ChartType: ChartTypeBar,
		Dims:      []string{"category"},
		Metrics:   []MetricConfig{{Field: "value", Agg: AggSum}},
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Data == nil {
		t.Error("Expected result, got nil")
	}
}
