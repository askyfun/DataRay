package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dataray/internal/datasource"
	"dataray/internal/model"
	"dataray/internal/query"

	"github.com/uptrace/bun"
)

type ChartService struct {
	db *bun.DB
}

func NewChartService(db *bun.DB) *ChartService {
	return &ChartService{db: db}
}

func (s *ChartService) GetByID(ctx context.Context, id int) (*model.Chart, error) {
	chart := &model.Chart{ID: id}
	if err := s.db.NewSelect().Model(chart).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("chart not found: %v", err)
	}
	return chart, nil
}

func (s *ChartService) List(ctx context.Context, limit, offset int) ([]model.Chart, error) {
	var charts []model.Chart
	query := s.db.NewSelect().Model(&charts)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return charts, nil
}

func (s *ChartService) Create(ctx context.Context, chart *model.Chart) (*model.Chart, error) {
	if chart.Config == "" {
		chart.Config = "{}"
	}
	if _, err := s.db.NewInsert().Model(chart).Exec(ctx); err != nil {
		return nil, err
	}
	return chart, nil
}

func (s *ChartService) Update(ctx context.Context, id int, name, chartType, config string) (*model.Chart, error) {
	chart := &model.Chart{ID: id}
	if err := s.db.NewSelect().Model(chart).WherePK().Scan(ctx); err != nil {
		return nil, err
	}

	chart.Name = name
	chart.ChartType = chartType
	chart.Config = config

	if _, err := s.db.NewUpdate().Model(chart).WherePK().Exec(ctx); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *ChartService) Delete(ctx context.Context, id int) error {
	if _, err := s.db.NewDelete().Model(&model.Chart{}).Where("id = ?", id).Exec(ctx); err != nil {
		return err
	}
	return nil
}

type ChartDataResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

func (s *ChartService) GetData(ctx context.Context, id int) (*ChartDataResult, error) {
	chart, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dataset, err := s.getDatasetByID(ctx, chart.DatasetID)
	if err != nil {
		return nil, err
	}

	datasourceModel, err := s.getDatasourceByID(ctx, dataset.DatasourceID)
	if err != nil {
		return nil, err
	}

	conn, err := s.getDatasourceConnection(ctx, datasourceModel)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var chartConfig map[string]interface{}
	json.Unmarshal([]byte(chart.Config), &chartConfig)

	sql, err := buildSQL(dataset, chartConfig)
	if err != nil {
		return nil, err
	}

	result, err := conn.Execute(ctx, sql)
	if err != nil {
		return nil, err
	}

	return &ChartDataResult{
		Columns: result.Columns,
		Rows:    result.Rows,
	}, nil
}

func (s *ChartService) ExecuteQuery(ctx context.Context, req *query.ChartQueryRequest) (query.ExecutorResult, error) {
	dataset, err := s.getDatasetByID(ctx, req.DatasetID)
	if err != nil {
		return query.ExecutorResult{}, err
	}

	datasourceModel, err := s.getDatasourceByID(ctx, dataset.DatasourceID)
	if err != nil {
		return query.ExecutorResult{}, err
	}

	driver, err := datasource.NewDriver(datasource.DriverType(datasourceModel.Type))
	if err != nil {
		return query.ExecutorResult{}, fmt.Errorf("failed to create driver: %v", err)
	}

	_ = driver

	conn, err := s.getDatasourceConnection(ctx, datasourceModel)
	if err != nil {
		return query.ExecutorResult{}, err
	}
	defer conn.Close()

	executor := query.NewExecutor(conn, dataset, datasourceModel)
	return executor.Execute(ctx, req)
}

func (s *ChartService) getDatasetByID(ctx context.Context, id int) (*model.Dataset, error) {
	dataset := &model.Dataset{ID: id}
	if err := s.db.NewSelect().Model(dataset).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("dataset not found: %v", err)
	}
	return dataset, nil
}

func (s *ChartService) getDatasourceByID(ctx context.Context, id int) (*model.Datasource, error) {
	ds := &model.Datasource{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %v", err)
	}
	return ds, nil
}

func (s *ChartService) getDatasourceConnection(ctx context.Context, ds *model.Datasource) (datasource.Connection, error) {
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

func parseDatasetColumns(columnsJSON string) ([]model.DatasetColumn, error) {
	if columnsJSON == "" || columnsJSON == "[]" {
		return nil, nil
	}
	var columns []model.DatasetColumn
	if err := json.Unmarshal([]byte(columnsJSON), &columns); err != nil {
		return nil, fmt.Errorf("failed to parse columns: %v", err)
	}
	return columns, nil
}

func resolveExpr(expr string, columns []model.DatasetColumn) (string, error) {
	if expr == "" {
		return "", nil
	}

	result := expr

	for i := 0; i < len(columns); i++ {
		col := columns[i]
		if col.Name == "" || col.Expr == "" {
			continue
		}

		pattern := "[" + col.Name + "]"
		if strings.Contains(result, pattern) {
			result = strings.ReplaceAll(result, pattern, col.Expr)
		}
	}

	return result, nil
}

func buildSQL(ds *model.Dataset, config map[string]interface{}) (string, error) {
	columns, err := parseDatasetColumns(ds.Columns)
	if err != nil {
		return "", err
	}

	xAxis, _ := config["xAxis"].(string)
	yAxis, _ := config["yAxis"].(string)
	chartType, _ := config["chartType"].(string)

	if xAxis == "" {
		return "", fmt.Errorf("xAxis is required")
	}

	xAxisExpr, err := resolveExpr(xAxis, columns)
	if err != nil {
		return "", fmt.Errorf("invalid xAxis: %v", err)
	}

	yAxisExpr := ""
	if yAxis != "" {
		yAxisExpr, err = resolveExpr(yAxis, columns)
		if err != nil {
			return "", fmt.Errorf("invalid yAxis: %v", err)
		}
	}

	var baseQuery string
	if ds.QueryType == "sql" && ds.QuerySQL.Valid {
		baseQuery = fmt.Sprintf("(%s) AS subq", ds.QuerySQL.String)
	} else if ds.TableName.Valid {
		if !isValidIdentifier(ds.TableName.String) {
			return "", fmt.Errorf("invalid table name")
		}
		baseQuery = ds.TableName.String
	} else {
		return "", fmt.Errorf("no table or query defined in dataset")
	}

	selectCols := xAxisExpr
	if yAxis != "" {
		if chartType == "pie" {
			selectCols = fmt.Sprintf("%s, SUM(%s) as value", xAxisExpr, yAxisExpr)
		} else {
			selectCols = fmt.Sprintf("%s, %s", xAxisExpr, yAxisExpr)
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", selectCols, baseQuery)

	if chartType == "pie" || chartType == "bar" {
		sql += fmt.Sprintf(" GROUP BY %s", xAxisExpr)
	}

	sql += fmt.Sprintf(" ORDER BY %s", xAxisExpr)
	sql += " LIMIT 1000"

	return sql, nil
}

func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	runes := []rune(name)
	first := runes[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return false
	}
	for _, c := range runes[1:] {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}
