package query

import (
	"encoding/json"
	"strings"
)

type ColumnInfo struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
	Role string `json:"role"`
	Type string `json:"type"`
}

type QueryAST struct {
	Source         string
	SourceType     SourceType
	Dimensions     []string
	Metrics        []MetricExpr
	Filters        []FilterExpr
	Sort           *SortExpr
	Pagination     *Pagination
	ColumnMappings map[string]string
}

type SourceType int

const (
	SourceTypeTable SourceType = iota
	SourceTypeSQL
)

type MetricExpr struct {
	Field     string
	FieldExpr string
	Agg       AggregationType
	Alias     string
	IsAgg     bool
}

type FilterExpr struct {
	Field     string
	FieldExpr string
	Op        FilterOperator
	Value     interface{}
	ValueEnd  interface{}
	Logic     string
}

type SortExpr struct {
	Field     string
	FieldExpr string
	Order     string
}

func (q *QueryAST) GetMetricFieldExpr(field string) string {
	if expr, ok := q.ColumnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

func (q *QueryAST) GetDimFieldExpr(field string) string {
	if expr, ok := q.ColumnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

func (q *QueryAST) GetFilterFieldExpr(field string) string {
	if expr, ok := q.ColumnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

func (q *QueryAST) GetSortFieldExpr(field string) string {
	if expr, ok := q.ColumnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

type QueryBuilder struct {
	columnMappings map[string]string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columnMappings: make(map[string]string),
	}
}

func (qb *QueryBuilder) WithColumnMappings(columns string) error {
	if columns == "" {
		return nil
	}

	var cols []ColumnInfo
	if err := json.Unmarshal([]byte(columns), &cols); err != nil {
		return err
	}

	for _, col := range cols {
		if col.Expr != "" && col.Name != "" {
			qb.columnMappings[col.Name] = col.Expr
		}
	}
	return nil
}

func (qb *QueryBuilder) Build(
	source string,
	sourceType SourceType,
	dims []string,
	metrics []MetricConfig,
	filters []FilterConfig,
	sort *SortConfig,
	pagination *Pagination,
) *QueryAST {
	ast := &QueryAST{
		Source:         source,
		SourceType:     sourceType,
		Dimensions:     dims,
		Filters:        make([]FilterExpr, 0, len(filters)),
		ColumnMappings: qb.columnMappings,
		Sort:           nil,
		Pagination:     pagination,
		Metrics:        make([]MetricExpr, 0, len(metrics)),
	}

	for _, m := range metrics {
		fieldExpr := qb.getFieldExpr(m.Field)
		expr := MetricExpr{
			Field:     m.Field,
			FieldExpr: fieldExpr,
			Agg:       m.Agg,
			Alias:     m.ResolveAlias(),
			IsAgg:     isAggregateFunction(fieldExpr),
		}
		ast.Metrics = append(ast.Metrics, expr)
	}

	for _, f := range filters {
		expr := FilterExpr{
			Field:     f.Field,
			FieldExpr: qb.getFieldExpr(f.Field),
			Op:        f.Op,
			Value:     f.Value,
			ValueEnd:  f.ValueEnd,
			Logic:     f.Logic,
		}
		ast.Filters = append(ast.Filters, expr)
	}

	if sort != nil {
		ast.Sort = &SortExpr{
			Field:     sort.Field,
			FieldExpr: qb.getFieldExpr(sort.Field),
			Order:     sort.Order,
		}
	}

	return ast
}

func (qb *QueryBuilder) getFieldExpr(field string) string {
	if expr, ok := qb.columnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

func isAggregateFunction(expr string) bool {
	upper := strings.ToUpper(expr)
	return strings.HasPrefix(upper, "COUNT") ||
		strings.HasPrefix(upper, "SUM") ||
		strings.HasPrefix(upper, "AVG") ||
		strings.HasPrefix(upper, "MIN") ||
		strings.HasPrefix(upper, "MAX") ||
		strings.HasPrefix(upper, "GROUP_CONCAT") ||
		strings.HasPrefix(upper, "JSON_ARRAYAGG") ||
		strings.HasPrefix(upper, "JSON_OBJECTAGG")
}
