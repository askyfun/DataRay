package service

import (
	"context"
	"testing"

	"dataray/internal/datasource"
)

type mockDriver struct {
	driverType datasource.DriverType
}

func (m *mockDriver) Type() datasource.DriverType {
	return m.driverType
}

func (m *mockDriver) Connect(ctx context.Context, config datasource.ConnectionConfig) (datasource.Connection, error) {
	return &mockConnection{}, nil
}

func (m *mockDriver) TestConnection(ctx context.Context, config datasource.ConnectionConfig) error {
	return nil
}

type mockConnection struct{}

func (m *mockConnection) Close() error                   { return nil }
func (m *mockConnection) Ping(ctx context.Context) error { return nil }
func (m *mockConnection) GetTables(ctx context.Context) ([]datasource.TableInfo, error) {
	return []datasource.TableInfo{
		{Name: "users", Comment: "users table"},
		{Name: "orders", Comment: "orders table"},
	}, nil
}
func (m *mockConnection) GetColumns(ctx context.Context, tableName string) ([]datasource.ColumnInfo, error) {
	return []datasource.ColumnInfo{
		{Name: "id", Type: "int", Comment: "primary key"},
		{Name: "name", Type: "varchar", Comment: "user name"},
	}, nil
}
func (m *mockConnection) Execute(ctx context.Context, sql string) (*datasource.QueryResult, error) {
	return &datasource.QueryResult{
		Columns: []string{"id", "name"},
		Rows: []map[string]interface{}{
			{"id": 1, "name": "test1"},
			{"id": 2, "name": "test2"},
		},
	}, nil
}

func TestGetByID_NotFound(t *testing.T) {
	t.Skip("Requires database connection - skipping")
}

func TestGetConnection_Timeout(t *testing.T) {
	t.Skip("Requires database connection - skipping")
}

func TestList_Empty(t *testing.T) {
	t.Skip("Requires database connection - skipping")
}

func TestCreate_Validation(t *testing.T) {
	t.Skip("Requires database connection - skipping")
}

func TestDelete_NotFound(t *testing.T) {
	t.Skip("Requires database connection - skipping")
}
