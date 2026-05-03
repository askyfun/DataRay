package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"dataray/internal/datasource"
	"dataray/internal/domain/entity"
	"dataray/internal/model"

	"github.com/uptrace/bun"
)

// Service defines the interface for datasource operations
type Service interface {
	// CRUD operations
	List(ctx context.Context, limit, offset int) ([]entity.Datasource, error)
	GetByID(ctx context.Context, id int) (*entity.Datasource, error)
	Create(ctx context.Context, ds *entity.Datasource) (*entity.Datasource, error)
	Update(ctx context.Context, ds *entity.Datasource) (*entity.Datasource, error)
	Delete(ctx context.Context, id int) error

	// Connection testing
	TestConnection(ctx context.Context, config entity.DatasourceConnectionConfig, driverType string) error

	// Schema operations
	GetTables(ctx context.Context, id int) ([]entity.TableInfo, error)
	GetColumns(ctx context.Context, id int, tableName string) ([]entity.ColumnInfo, error)

	// Preview operations
	Preview(ctx context.Context, id int, tableName, querySQL, queryType string) (*entity.PreviewResult, error)
	GetFieldDistribution(ctx context.Context, id int, tableName, querySQL, queryType, fieldName string, limit int) (*entity.FieldDistribution, error)
}

// datasourceService implements the Service interface
type datasourceService struct {
	db *bun.DB
}

// NewService creates a new datasource service
func NewService(db *bun.DB) Service {
	return &datasourceService{db: db}
}

// List returns all datasources with pagination
func (s *datasourceService) List(ctx context.Context, limit, offset int) ([]entity.Datasource, error) {
	var datasources []model.Datasource
	q := s.db.NewSelect().Model(&datasources)
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to list datasources: %w", err)
	}
	return toEntityList(datasources), nil
}

// GetByID returns a datasource by ID
func (s *datasourceService) GetByID(ctx context.Context, id int) (*entity.Datasource, error) {
	ds := &model.Datasource{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}
	return toEntity(ds), nil
}

// Create creates a new datasource
func (s *datasourceService) Create(ctx context.Context, ds *entity.Datasource) (*entity.Datasource, error) {
	m := toModel(ds)
	if _, err := s.db.NewInsert().Model(m).Returning("*").Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}
	return toEntity(m), nil
}

// Update updates an existing datasource
func (s *datasourceService) Update(ctx context.Context, ds *entity.Datasource) (*entity.Datasource, error) {
	m := toModel(ds)
	if _, err := s.db.NewUpdate().Model(m).WherePK().Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to update datasource: %w", err)
	}
	updated := &model.Datasource{ID: ds.ID}
	if err := s.db.NewSelect().Model(updated).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get updated datasource: %w", err)
	}
	return toEntity(updated), nil
}

// Delete deletes a datasource by ID
func (s *datasourceService) Delete(ctx context.Context, id int) error {
	if _, err := s.db.NewDelete().Model(&model.Datasource{}).Where("id = ?", id).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete datasource: %w", err)
	}
	return nil
}

// TestConnection tests the connection to a datasource
func (s *datasourceService) TestConnection(ctx context.Context, config entity.DatasourceConnectionConfig, driverType string) error {
	driver, err := datasource.NewDriver(datasource.DriverType(driverType))
	if err != nil {
		return fmt.Errorf("unsupported driver type: %s", driverType)
	}

	connConfig := datasource.ConnectionConfig{
		Host:         config.Host,
		Port:         config.Port,
		DatabaseName: config.DatabaseName,
		Username:     config.Username,
		Password:     config.Password,
	}

	if err := driver.TestConnection(ctx, connConfig); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	return nil
}

// GetTables returns all tables from a datasource
func (s *datasourceService) GetTables(ctx context.Context, id int) ([]entity.TableInfo, error) {
	ds, err := s.getDatasourceModel(ctx, id)
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(ctx, ds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tables, err := conn.GetTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	return toTableInfoList(tables), nil
}

// GetColumns returns all columns for a table
func (s *datasourceService) GetColumns(ctx context.Context, id int, tableName string) ([]entity.ColumnInfo, error) {
	ds, err := s.getDatasourceModel(ctx, id)
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(ctx, ds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	columns, err := conn.GetColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	return toColumnInfoList(columns), nil
}

// Preview returns preview data from a datasource
func (s *datasourceService) Preview(ctx context.Context, id int, tableName, querySQL, queryType string) (*entity.PreviewResult, error) {
	ds, err := s.getDatasourceModel(ctx, id)
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(ctx, ds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sql string
	if queryType == "sql" {
		sql = fmt.Sprintf("%s LIMIT 10", querySQL)
	} else {
		sql = fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableName)
	}

	result, err := conn.Execute(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return &entity.PreviewResult{
		Columns: result.Columns,
		Data:    result.Rows,
	}, nil
}

// GetFieldDistribution returns field value distribution
func (s *datasourceService) GetFieldDistribution(ctx context.Context, id int, tableName, querySQL, queryType, fieldName string, limit int) (*entity.FieldDistribution, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	ds, err := s.getDatasourceModel(ctx, id)
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(ctx, ds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sql string
	if queryType == "sql" {
		sql = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM (%s) as _subquery GROUP BY %s ORDER BY _count DESC LIMIT %d",
			fieldName, querySQL, fieldName, limit)
	} else {
		sql = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM %s GROUP BY %s ORDER BY _count DESC LIMIT %d",
			fieldName, tableName, fieldName, limit)
	}

	result, err := conn.Execute(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	// Get total count
	var totalQuery string
	if queryType == "sql" {
		totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM (%s) as _subquery", querySQL)
	} else {
		totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM %s", tableName)
	}

	var totalCount int64
	totalResult, err := conn.Execute(ctx, totalQuery)
	if err == nil && len(totalResult.Rows) > 0 {
		if val, ok := totalResult.Rows[0]["_total"]; ok {
			switch v := val.(type) {
			case int64:
				totalCount = v
			case float64:
				totalCount = int64(v)
			}
		}
	}

	// Format distribution
	distribution := make([]entity.FieldValueCount, 0, len(result.Rows))
	for _, row := range result.Rows {
		value := row[fieldName]
		count := row["_count"]

		var countInt int64
		switch v := count.(type) {
		case int64:
			countInt = v
		case float64:
			countInt = int64(v)
		}

		var percentage float64
		if totalCount > 0 {
			percentage = math.Round(float64(countInt)*10000/float64(totalCount)) / 100
		}

		distribution = append(distribution, entity.FieldValueCount{
			Value:      value,
			Count:      countInt,
			Percentage: percentage,
		})
	}

	return &entity.FieldDistribution{
		FieldName:    fieldName,
		TotalCount:   totalCount,
		UniqueCount:  len(result.Rows),
		Distribution: distribution,
	}, nil
}

// Helper functions

func (s *datasourceService) getDatasourceModel(ctx context.Context, id int) (*model.Datasource, error) {
	ds := &model.Datasource{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}
	return ds, nil
}

func (s *datasourceService) connect(ctx context.Context, ds *model.Datasource) (datasource.Connection, error) {
	driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
	if err != nil {
		return nil, fmt.Errorf("unsupported driver type: %s", ds.Type)
	}

	config := datasource.ConnectionConfig{
		Host:         ds.Host,
		Port:         ds.Port,
		DatabaseName: ds.DatabaseName,
		Username:     ds.Username,
		Password:     ds.Password,
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := driver.Connect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	return conn, nil
}

// Conversion functions

func toEntity(m *model.Datasource) *entity.Datasource {
	e := &entity.Datasource{
		ID:           m.ID,
		Name:         m.Name,
		Type:         m.Type,
		Host:         m.Host,
		Port:         m.Port,
		DatabaseName: m.DatabaseName,
		Username:     m.Username,
		Password:     m.Password,
	}
	if m.CreatedAt.Valid {
		e.CreatedAt = m.CreatedAt.Time.Format(time.RFC3339)
	}
	if m.UpdatedAt.Valid {
		e.UpdatedAt = m.UpdatedAt.Time.Format(time.RFC3339)
	}
	return e
}

func toEntityList(models []model.Datasource) []entity.Datasource {
	result := make([]entity.Datasource, len(models))
	for i, m := range models {
		e := toEntity(&m)
		result[i] = *e
	}
	return result
}

func toModel(e *entity.Datasource) *model.Datasource {
	m := &model.Datasource{
		ID:           e.ID,
		Name:         e.Name,
		Type:         e.Type,
		Host:         e.Host,
		Port:         e.Port,
		DatabaseName: e.DatabaseName,
		Username:     e.Username,
		Password:     e.Password,
	}
	if e.CreatedAt != "" {
		t, _ := time.Parse(time.RFC3339, e.CreatedAt)
		m.CreatedAt = sql.NullTime{Time: t, Valid: true}
	}
	if e.UpdatedAt != "" {
		t, _ := time.Parse(time.RFC3339, e.UpdatedAt)
		m.UpdatedAt = sql.NullTime{Time: t, Valid: true}
	}
	return m
}

func toTableInfoList(tables []datasource.TableInfo) []entity.TableInfo {
	result := make([]entity.TableInfo, len(tables))
	for i, t := range tables {
		result[i] = entity.TableInfo{Name: t.Name, Comment: t.Comment}
	}
	return result
}

func toColumnInfoList(columns []datasource.ColumnInfo) []entity.ColumnInfo {
	result := make([]entity.ColumnInfo, len(columns))
	for i, c := range columns {
		result[i] = entity.ColumnInfo{Name: c.Name, DataType: c.Type, Comment: c.Comment}
	}
	return result
}
