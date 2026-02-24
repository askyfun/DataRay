package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"dataray/internal/datasource"
	"dataray/internal/model"

	"github.com/uptrace/bun"
)

type DatasetService struct {
	db *bun.DB
}

func NewDatasetService(db *bun.DB) *DatasetService {
	return &DatasetService{db: db}
}

func (s *DatasetService) GetByID(ctx context.Context, id int) (*model.Dataset, error) {
	ds := &model.Dataset{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("dataset not found: %v", err)
	}
	return ds, nil
}

func (s *DatasetService) List(ctx context.Context, limit, offset int) ([]model.Dataset, error) {
	var datasets []model.Dataset
	query := s.db.NewSelect().Model(&datasets)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return datasets, nil
}

func (s *DatasetService) Create(ctx context.Context, ds *model.Dataset) (*model.Dataset, error) {
	if ds.QueryType == "" {
		ds.QueryType = "table"
	}
	if ds.Mode == "" {
		ds.Mode = "direct"
	}
	if ds.Tags == "" {
		ds.Tags = "[]"
	}
	if ds.QualityRules == "" {
		ds.QualityRules = "[]"
	}
	if ds.Columns == "" {
		ds.Columns = "[]"
	}
	ds.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if _, err := s.db.NewInsert().Model(ds).Exec(ctx); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasetService) Delete(ctx context.Context, id int) error {
	if _, err := s.db.NewDelete().Model(&model.Dataset{}).Where("id = ?", id).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *DatasetService) UpdateColumns(ctx context.Context, id int, columnsJSON string) (*model.Dataset, error) {
	ds := &model.Dataset{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, err
	}

	ds.Columns = columnsJSON
	if _, err := s.db.NewUpdate().Model(ds).WherePK().Exec(ctx); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *DatasetService) GetColumns(ctx context.Context, id int) ([]model.DatasetColumn, error) {
	ds, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ds.Columns != "" && ds.Columns != "[]" {
		var savedColumns []model.DatasetColumn
		if err := json.Unmarshal([]byte(ds.Columns), &savedColumns); err == nil && len(savedColumns) > 0 {
			return savedColumns, nil
		}
	}

	datasourceModel, err := s.getDatasourceByID(ctx, ds.DatasourceID)
	if err != nil {
		return nil, err
	}

	conn, err := s.getDatasourceConnection(ctx, datasourceModel)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var tableName string
	if ds.QueryType == "sql" {
		tableName = fmt.Sprintf("(%s) as subq", ds.QuerySQL.String)
	} else {
		tableName = ds.TableName.String
	}

	dbColumns, err := conn.GetColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	mapper, _ := model.NewDataTypeMapper("starrocks")
	result := make([]model.DatasetColumn, len(dbColumns))

	for i, col := range dbColumns {
		stdType, typeConfig, _ := mapper.ToStandard(col.Type)
		result[i] = model.DatasetColumn{
			Name:       col.Name,
			Expr:       "`" + col.Name + "`",
			Type:       stdType,
			TypeConfig: typeConfig,
			Comment:    "",
			Role:       inferRole(string(stdType)),
		}
	}
	return result, nil
}

func (s *DatasetService) Preview(ctx context.Context, id int) (*datasource.QueryResult, error) {
	ds, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	datasourceModel, err := s.getDatasourceByID(ctx, ds.DatasourceID)
	if err != nil {
		return nil, err
	}

	conn, err := s.getDatasourceConnection(ctx, datasourceModel)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var tableName string
	if ds.QueryType == "sql" {
		tableName = fmt.Sprintf("(%s) as subq", ds.QuerySQL.String)
	} else {
		tableName = ds.TableName.String
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableName)
	result, err := conn.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}

	return result, nil
}

func (s *DatasetService) GetDatasourceID(ctx context.Context, id int) (int, error) {
	ds, err := s.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return ds.DatasourceID, nil
}

func (s *DatasetService) GetDatasource(ctx context.Context, id int) (*model.Datasource, error) {
	ds, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.getDatasourceByID(ctx, ds.DatasourceID)
}

func (s *DatasetService) getDatasourceByID(ctx context.Context, id int) (*model.Datasource, error) {
	ds := &model.Datasource{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %v", err)
	}
	return ds, nil
}

func (s *DatasetService) getDatasourceConnection(ctx context.Context, ds *model.Datasource) (datasource.Connection, error) {
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

	ctx, cancel := context.WithTimeout(ctx, 30)
	defer cancel()

	conn, err := driver.Connect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return conn, nil
}

func inferRole(dataType string) string {
	lowerType := dataType
	if lowerType == "int" || lowerType == "integer" ||
		lowerType == "float" || lowerType == "double" ||
		lowerType == "decimal" || lowerType == "numeric" ||
		lowerType == "real" || lowerType == "number" {
		return "metric"
	}
	return "dimension"
}
