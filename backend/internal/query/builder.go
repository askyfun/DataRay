package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Builder Query 构建器
type Builder struct {
	baseQuery      string
	dims           []string
	metrics        []MetricConfig
	filters        []FilterConfig
	sort           *SortConfig
	pagination     *Pagination
	columnMappings map[string]string
}

// NewBuilder 创建新的 QueryBuilder
func NewBuilder(baseQuery string) *Builder {
	return &Builder{
		baseQuery:      baseQuery,
		columnMappings: make(map[string]string),
	}
}

// WithColumnMappings 设置列映射 (字段名 -> expr)
func (b *Builder) WithColumnMappings(columns string) error {
	if columns == "" {
		return nil
	}

	var cols []ColumnInfo
	if err := json.Unmarshal([]byte(columns), &cols); err != nil {
		return err
	}

	for _, col := range cols {
		if col.Expr != "" && col.Name != "" {
			b.columnMappings[col.Name] = col.Expr
		}
	}
	return nil
}

// WithDims 设置维度
func (b *Builder) WithDims(dims []string) *Builder {
	b.dims = dims
	return b
}

// WithMetrics 设置指标
func (b *Builder) WithMetrics(metrics []MetricConfig) *Builder {
	b.metrics = metrics
	return b
}

// WithFilters 设置过滤条件
func (b *Builder) WithFilters(filters []FilterConfig) *Builder {
	b.filters = filters
	return b
}

// WithSort 设置排序
func (b *Builder) WithSort(sort *SortConfig) *Builder {
	b.sort = sort
	return b
}

// WithPagination 设置分页
func (b *Builder) WithPagination(pagination *Pagination) *Builder {
	b.pagination = pagination
	return b
}

// Build 生成基础查询 SQL (不带分页)
func (b *Builder) Build() string {
	var sb strings.Builder

	// SELECT 子句
	sb.WriteString("SELECT ")
	sb.WriteString(b.buildSelectClause())

	// FROM 子句 - 根据 baseQuery 类型决定是否使用子查询
	if b.isSQLQuery(b.baseQuery) {
		sb.WriteString(" FROM (")
		sb.WriteString(b.baseQuery)
		sb.WriteString(") AS _t0")
	} else {
		sb.WriteString(" FROM ")
		sb.WriteString(b.baseQuery)
	}

	// WHERE 子句
	whereClause := b.buildWhereClause()
	if whereClause != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(whereClause)
	}

	// GROUP BY 子句
	groupByClause := b.buildGroupByClause()
	if groupByClause != "" {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(groupByClause)
	}

	// ORDER BY 子句 (仅用于非分页查询)
	if b.sort != nil && b.pagination == nil {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.sort.Field)
		sb.WriteString(" ")
		sb.WriteString(b.sort.Order)
	}

	return sb.String()
}

// isSQLQuery 判断 baseQuery 是否为 SQL 查询
func (b *Builder) isSQLQuery(baseQuery string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(baseQuery))
	return strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "WITH") ||
		strings.Contains(trimmed, " JOIN ")
}

// BuildWithPagination 生成分页查询 SQL
func (b *Builder) BuildWithPagination() (dataSQL, countSQL string) {
	baseSQL := b.Build()

	// 数据查询 SQL
	var sb strings.Builder
	sb.WriteString(baseSQL)

	// ORDER BY 子句 (分页查询需要)
	if b.sort != nil {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.sort.Field)
		sb.WriteString(" ")
		sb.WriteString(b.sort.Order)
	}

	// LIMIT 子句
	if b.pagination != nil {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(b.pagination.PageSize))
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.Itoa((b.pagination.Page - 1) * b.pagination.PageSize))
	}

	dataSQL = sb.String()

	// COUNT 查询 SQL
	var countSb strings.Builder
	countSb.WriteString("SELECT COUNT(*) AS _total FROM (")
	countSb.WriteString(baseSQL)
	countSb.WriteString(") AS _count_query")

	countSQL = countSb.String()

	return
}

// BuildDataQuery 仅生成数据查询 SQL
func (b *Builder) BuildDataQuery() string {
	var sb strings.Builder
	sb.WriteString(b.Build())

	if b.sort != nil {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.sort.Field)
		sb.WriteString(" ")
		sb.WriteString(b.sort.Order)
	}

	if b.pagination != nil {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(b.pagination.PageSize))
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.Itoa((b.pagination.Page - 1) * b.pagination.PageSize))
	}

	return sb.String()
}

// BuildCountQuery 生成 COUNT 查询 SQL
func (b *Builder) BuildCountQuery() string {
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(*) AS _total FROM (")
	sb.WriteString(b.Build())
	sb.WriteString(") AS _count_query")
	return sb.String()
}

// buildSelectClause 构建 SELECT 子句
func (b *Builder) buildSelectClause() string {
	var parts []string

	for _, dim := range b.dims {
		parts = append(parts, dim)
	}

	for _, metric := range b.metrics {
		aggFunc := metric.Agg.GetAggFunc()
		alias := metric.ResolveAlias()

		fieldExpr := metric.Field
		if expr, ok := b.columnMappings[metric.Field]; ok && expr != "" {
			fieldExpr = expr
		}

		expr := fmt.Sprintf("%s(%s) AS %s", aggFunc, fieldExpr, alias)
		parts = append(parts, expr)
	}

	if len(parts) == 0 {
		return "*"
	}

	return strings.Join(parts, ", ")
}

// buildWhereClause 构建 WHERE 子句
func (b *Builder) buildWhereClause() string {
	if len(b.filters) == 0 {
		return ""
	}

	var conditions []string
	for _, filter := range b.filters {
		condition := buildFilterCondition(filter)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ")
}

// buildGroupByClause 构建 GROUP BY 子句
func (b *Builder) buildGroupByClause() string {
	if len(b.dims) == 0 {
		return ""
	}
	return strings.Join(b.dims, ", ")
}

// buildFilterCondition 构建单个过滤条件
func buildFilterCondition(filter FilterConfig) string {
	field := filter.Field
	op := filter.Op.ToString()

	switch filter.Op {
	case FilterIsNull, FilterIsNotNull:
		return fmt.Sprintf("%s %s", field, op)

	case FilterIn:
		formatValue := FormatValue(filter.Op, filter.Value)
		return fmt.Sprintf("%s IN %s", field, formatValue)

	case FilterBetween:
		return fmt.Sprintf("%s BETWEEN '%v' AND '%v'", field, filter.Value, filter.ValueEnd)

	case FilterLike:
		return fmt.Sprintf("%s LIKE '%%%v%%'", field, filter.Value)

	default:
		formatValue := FormatValue(filter.Op, filter.Value)
		return fmt.Sprintf("%s %s %s", field, op, formatValue)
	}
}
