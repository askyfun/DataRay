package query

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"dataray/internal/datasource"
	"dataray/internal/model"
)

type GeneratedSQL struct {
	Select string `json:"select_sql"`
	Count  string `json:"count_sql,omitempty"`
}

type ExecutorResult struct {
	Data interface{} `json:"data"`
	GeneratedSQL
}

// Executor 查询执行器
type Executor struct {
	conn       datasource.Connection
	dataset    *model.Dataset
	datasource *model.Datasource
}

// NewExecutor 创建新的执行器
func NewExecutor(conn datasource.Connection, dataset *model.Dataset, ds *model.Datasource) *Executor {
	return &Executor{
		conn:       conn,
		dataset:    dataset,
		datasource: ds,
	}
}

// Execute 执行图表查询
func (e *Executor) Execute(ctx context.Context, req *ChartQueryRequest) (ExecutorResult, error) {
	baseQuery, sourceType := e.getBaseQuery()
	if baseQuery == "" {
		return ExecutorResult{}, fmt.Errorf("dataset has no valid query_sql or table_name")
	}

	dialect := ParseDialect(e.datasource.Type)

	qb := NewQueryBuilder()
	qb.WithColumnMappings(e.dataset.Columns)

	ast := qb.Build(
		baseQuery,
		sourceType,
		req.Dims,
		req.Metrics,
		req.Filters,
		req.Sort,
		req.Pagination,
	)

	sql, countSQL := BuildQueryString(dialect, ast)
	slog.Debug("generated SQL", "select", sql, "count", countSQL)

	processor := GetProcessor(req.ChartType)

	if req.ChartType == ChartTypeTable && req.Pagination != nil {
		slog.Debug("executing data query", "sql", sql)

		result, err := e.conn.Execute(ctx, sql)
		if err != nil {
			return ExecutorResult{}, fmt.Errorf("query failed: %v", err)
		}

		total := len(result.Rows)
		if req.Pagination.PageSize > 0 && len(result.Rows) >= req.Pagination.PageSize && countSQL != "" {
			slog.Debug("executing count query", "sql", countSQL)

			countResult, err := e.conn.Execute(ctx, countSQL)
			if err == nil && len(countResult.Rows) > 0 {
				if totalVal, ok := countResult.Rows[0]["_total"]; ok {
					switch v := totalVal.(type) {
					case int64:
						total = int(v)
					case float64:
						total = int(v)
					}
				}
			}
		}

		rows := result.Rows
		page := req.Pagination.Page
		pageSize := req.Pagination.PageSize
		totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

		start := (page - 1) * pageSize
		end := start + pageSize
		if start > len(rows) {
			rows = []map[string]any{}
		} else {
			if end > len(rows) {
				end = len(rows)
			}
			rows = rows[start:end]
		}

		columns := []string{}

		// 按维度在前、指标在后的顺序构建 columns
		for _, dim := range req.Dims {
			columns = append(columns, dim)
		}
		for _, metric := range req.Metrics {
			alias := metric.ResolveAlias()
			columns = append(columns, alias)
		}

		// 重新排序数据行的字段顺序，使其与 columns 一致
		orderedRows := make([]map[string]any, len(rows))
		for i, row := range rows {
			orderedRow := make(map[string]any)
			for _, col := range columns {
				if val, ok := row[col]; ok {
					orderedRow[col] = val
				}
			}
			orderedRows[i] = orderedRow
		}

		return ExecutorResult{
			Data: &TableResponse{
				Columns: columns,
				Data:    orderedRows,
				Pagination: TablePagination{
					Page:       page,
					PageSize:   pageSize,
					Total:      total,
					TotalPages: totalPages,
				},
			},
			GeneratedSQL: GeneratedSQL{
				Select: sql,
				Count:  countSQL,
			},
		}, nil
	}

	slog.Debug("executing chart query", "sql", sql)

	result, err := e.conn.Execute(ctx, sql)
	if err != nil {
		return ExecutorResult{}, fmt.Errorf("query failed: %v", err)
	}

	data, err := processor.Process(result.Rows, req.Dims, req.Metrics)
	if err != nil {
		return ExecutorResult{}, fmt.Errorf("process failed: %v", err)
	}

	return ExecutorResult{
		Data: data,
		GeneratedSQL: GeneratedSQL{
			Select: sql,
		},
	}, nil
}

// getBaseQuery 获取基础查询 SQL
func (e *Executor) getBaseQuery() (string, SourceType) {
	if e.dataset.QueryType == "sql" && e.dataset.QuerySQL.Valid {
		return e.dataset.QuerySQL.String, SourceTypeSQL
	}
	if e.dataset.TableName.Valid {
		return e.dataset.TableName.String, SourceTypeTable
	}
	return "", SourceTypeTable
}

// ExecuteRawQuery 执行原始查询
func (e *Executor) ExecuteRawQuery(ctx context.Context, sql string) ([]map[string]any, error) {
	result, err := e.conn.Execute(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return result.Rows, nil
}

// Close 关闭连接
func (e *Executor) Close() error {
	return e.conn.Close()
}
