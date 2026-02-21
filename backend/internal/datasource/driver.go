package datasource

import (
	"context"
)

// DriverType represents the type of database driver
type DriverType string

const (
	DriverPostgreSQL DriverType = "postgresql"
	DriverClickHouse DriverType = "clickhouse"
	DriverMySQL      DriverType = "mysql"
	DriverStarRocks  DriverType = "starrocks"
)

// ConnectionConfig holds the connection configuration for a datasource
type ConnectionConfig struct {
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string
}

// TableInfo represents a table in the database
type TableInfo struct {
	Name    string
	Comment string
}

// ColumnInfo represents a column in a table
type ColumnInfo struct {
	Name    string
	Type    string
	Comment string
}

// QueryResult represents the result of a query
type QueryResult struct {
	Columns []string
	Rows    []map[string]interface{}
}

// Driver is the interface that all database drivers must implement
type Driver interface {
	// Type returns the driver type
	Type() DriverType

	// Connect establishes a connection to the database
	Connect(ctx context.Context, config ConnectionConfig) (Connection, error)

	// TestConnection tests if the connection is valid
	TestConnection(ctx context.Context, config ConnectionConfig) error
}

// Connection represents a database connection
type Connection interface {
	// Close closes the connection
	Close() error

	// Ping pings the database to check connection
	Ping(ctx context.Context) error

	// GetTables returns the list of tables in the database
	GetTables(ctx context.Context) ([]TableInfo, error)

	// GetColumns returns the list of columns for a table
	GetColumns(ctx context.Context, tableName string) ([]ColumnInfo, error)

	// Execute executes a query and returns the result
	Execute(ctx context.Context, sql string) (*QueryResult, error)
}

// NewDriver creates a new driver instance based on the driver type
func NewDriver(driverType DriverType) (Driver, error) {
	switch driverType {
	case DriverPostgreSQL:
		return NewPostgreSQLDriver(), nil
	case DriverClickHouse:
		return NewClickHouseDriver(), nil
	case DriverMySQL:
		return NewMySQLDriver(), nil
	case DriverStarRocks:
		return NewStarRocksDriver(), nil
	default:
		return nil, nil
	}
}
