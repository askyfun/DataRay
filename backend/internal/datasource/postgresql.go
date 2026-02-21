package datasource

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresqlDriver struct{}

func NewPostgreSQLDriver() Driver {
	return &postgresqlDriver{}
}

func (d *postgresqlDriver) Type() DriverType {
	return DriverPostgreSQL
}

func (d *postgresqlDriver) Connect(ctx context.Context, config ConnectionConfig) (Connection, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.DatabaseName)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &postgresqlConnection{pool: pool}, nil
}

func (d *postgresqlDriver) TestConnection(ctx context.Context, config ConnectionConfig) error {
	conn, err := d.Connect(ctx, config)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.Ping(ctx)
}

type postgresqlConnection struct {
	pool *pgxpool.Pool
}

func (c *postgresqlConnection) Close() error {
	c.pool.Close()
	return nil
}

func (c *postgresqlConnection) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

func (c *postgresqlConnection) GetTables(ctx context.Context) ([]TableInfo, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT t.table_name, COALESCE(pgtd.description, '') as table_comment
		FROM information_schema.tables t
		LEFT JOIN pg_class pc ON pc.relname = t.table_name AND pc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)
		LEFT JOIN pg_description pgtd ON pc.oid = pgtd.objoid AND pgtd.objsubid = 0
		WHERE t.table_schema = 'public'
		ORDER BY t.table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var name, comment string
		if err := rows.Scan(&name, &comment); err != nil {
			return nil, err
		}
		tables = append(tables, TableInfo{Name: name, Comment: comment})
	}
	return tables, nil
}

func (c *postgresqlConnection) GetColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf(`
		SELECT c.column_name, c.data_type, COALESCE(pgcd.description, '') as column_comment
		FROM information_schema.columns c
		LEFT JOIN pg_class pc ON pc.relname = c.table_name AND pc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = c.table_schema)
		LEFT JOIN pg_description pgcd ON pc.oid = pgcd.objoid AND pgcd.objsubid = c.ordinal_position
		WHERE c.table_name = '%s' AND c.table_schema = 'public'
		ORDER BY c.ordinal_position`, tableName)

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var name, colType, comment string
		if err := rows.Scan(&name, &colType, &comment); err != nil {
			return nil, err
		}
		columns = append(columns, ColumnInfo{Name: name, Type: colType, Comment: comment})
	}
	return columns, nil
}

func (c *postgresqlConnection) Execute(ctx context.Context, sql string) (*QueryResult, error) {
	rows, err := c.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = string(fd.Name)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return &QueryResult{
		Columns: columns,
		Rows:    results,
	}, nil
}
