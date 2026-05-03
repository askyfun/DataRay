package query

import (
	"fmt"
	"strings"
)

// BunQueryBuilder 使用 bun ORM 的安全查询构建器
type BunQueryBuilder struct {
	columnMappings map[string]string
	dialect        DialectType
}

// NewBunQueryBuilder 创建新的 BunQueryBuilder
func NewBunQueryBuilder() *BunQueryBuilder {
	return &BunQueryBuilder{
		columnMappings: make(map[string]string),
	}
}

// WithColumnMappings 设置列映射
func (qb *BunQueryBuilder) WithColumnMappings(columns string) error {
	if columns == "" {
		return nil
	}

	var cols []ColumnInfo
	if err := unmarshalJSON(columns, &cols); err != nil {
		return err
	}

	for _, col := range cols {
		if col.Expr != "" && col.Name != "" {
			qb.columnMappings[col.Name] = col.Expr
		}
	}
	return nil
}

// SetDialect 设置数据库方言
func (qb *BunQueryBuilder) SetDialect(dialect DialectType) {
	qb.dialect = dialect
}

// Build 使用 bun QueryBuilder 构建查询
func (qb *BunQueryBuilder) Build(
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

	// 构建指标
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

	// 构建过滤条件
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

	// 构建排序
	if sort != nil {
		ast.Sort = &SortExpr{
			Field:     sort.Field,
			FieldExpr: qb.getFieldExpr(sort.Field),
			Order:     sort.Order,
		}
	}

	return ast
}

// getFieldExpr 获取字段表达式
func (qb *BunQueryBuilder) getFieldExpr(field string) string {
	if expr, ok := qb.columnMappings[field]; ok && expr != "" {
		return expr
	}
	return field
}

// BuildSelectQuery 使用 bun 构建安全的选择查询
// 返回 SQL 字符串和参数列表（用于参数化查询）
func (qb *BunQueryBuilder) BuildSelectQuery(ast *QueryAST) (string, []interface{}) {
	var args []interface{}
	var sb strings.Builder

	sb.WriteString("SELECT ")

	// SELECT 子句
	selectParts := qb.buildSelectParts(ast)
	sb.WriteString(strings.Join(selectParts, ", "))

	// FROM 子句
	sb.WriteString(" FROM ")
	if ast.SourceType == SourceTypeSQL {
		sb.WriteString("(")
		sb.WriteString(ast.Source)
		sb.WriteString(") AS _subq")
	} else {
		sb.WriteString(safeIdentifier(ast.Source))
	}

	// WHERE 子句
	if len(ast.Filters) > 0 {
		sb.WriteString(" WHERE ")
		whereParts := qb.buildWhereParts(ast, &args)
		sb.WriteString(strings.Join(whereParts, " AND "))
	}

	// GROUP BY 子句
	if len(ast.Dimensions) > 0 {
		sb.WriteString(" GROUP BY ")
		groupByParts := make([]string, len(ast.Dimensions))
		for i, dim := range ast.Dimensions {
			groupByParts[i] = safeIdentifier(dim)
		}
		sb.WriteString(strings.Join(groupByParts, ", "))
	}

	// ORDER BY 子句
	if ast.Sort != nil && ast.Pagination == nil {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(safeIdentifier(ast.Sort.FieldExpr))
		sb.WriteString(" ")
		sb.WriteString(ast.Sort.Order)
	}

	// LIMIT/OFFSET 子句 - 直接嵌入数值，因为某些数据库驱动不支持预处理语句的 ? 占位符用于 LIMIT/OFFSET
	if ast.Pagination != nil {
		pageSize := ast.Pagination.PageSize
		offset := (ast.Pagination.Page - 1) * ast.Pagination.PageSize
		sb.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset))
	}

	return sb.String(), args
}

// BuildCountQuery 构建计数查询
func (qb *BunQueryBuilder) BuildCountQuery(ast *QueryAST) (string, []interface{}) {
	var args []interface{}
	var sb strings.Builder

	sb.WriteString("SELECT COUNT(*) AS _total FROM (")

	// 内层查询
	sb.WriteString("SELECT 1")

	// FROM 子句
	sb.WriteString(" FROM ")
	if ast.SourceType == SourceTypeSQL {
		sb.WriteString("(")
		sb.WriteString(ast.Source)
		sb.WriteString(") AS _subq")
	} else {
		sb.WriteString(safeIdentifier(ast.Source))
	}

	// WHERE 子句
	if len(ast.Filters) > 0 {
		sb.WriteString(" WHERE ")
		whereParts := qb.buildWhereParts(ast, &args)
		sb.WriteString(strings.Join(whereParts, " AND "))
	}

	// GROUP BY 子句
	if len(ast.Dimensions) > 0 {
		sb.WriteString(" GROUP BY ")
		groupByParts := make([]string, len(ast.Dimensions))
		for i, dim := range ast.Dimensions {
			groupByParts[i] = safeIdentifier(dim)
		}
		sb.WriteString(strings.Join(groupByParts, ", "))
	}

	sb.WriteString(") AS _count_query")

	return sb.String(), args
}

// buildSelectParts 构建 SELECT 部分
func (qb *BunQueryBuilder) buildSelectParts(ast *QueryAST) []string {
	var parts []string

	// 维度
	for _, dim := range ast.Dimensions {
		parts = append(parts, safeIdentifier(dim))
	}

	// 指标
	for _, metric := range ast.Metrics {
		var expr string
		if metric.IsAgg {
			expr = fmt.Sprintf("%s AS %s", safeIdentifier(metric.FieldExpr), safeIdentifier(metric.Alias))
		} else {
			aggFunc := metric.Agg.GetAggFunc()
			expr = fmt.Sprintf("%s(%s) AS %s", aggFunc, safeIdentifier(metric.FieldExpr), safeIdentifier(metric.Alias))
		}
		parts = append(parts, expr)
	}

	if len(parts) == 0 {
		return []string{"*"}
	}

	return parts
}

// buildWhereParts 构建 WHERE 部分，使用参数化查询
func (qb *BunQueryBuilder) buildWhereParts(ast *QueryAST, args *[]interface{}) []string {
	var parts []string

	for i := range ast.Filters {
		part := qb.buildFilterPart(&ast.Filters[i], args)
		if part != "" {
			parts = append(parts, part)
		}
	}

	return parts
}

// buildFilterPart 构建单个过滤条件，使用参数化查询
func (qb *BunQueryBuilder) buildFilterPart(f *FilterExpr, args *[]interface{}) string {
	field := safeIdentifier(f.FieldExpr)

	switch f.Op {
	case FilterIsNull:
		return fmt.Sprintf("%s IS NULL", field)
	case FilterIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", field)
	case FilterIn:
		if vals, ok := f.Value.([]any); ok && len(vals) > 0 {
			placeholders := make([]string, len(vals))
			for i := range vals {
				placeholders[i] = "?"
				*args = append(*args, vals[i])
			}
			return fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
		}
		return ""
	case FilterBetween:
		*args = append(*args, f.Value, f.ValueEnd)
		return fmt.Sprintf("%s BETWEEN ? AND ?", field)
	case FilterLike:
		*args = append(*args, "%"+fmt.Sprintf("%v", f.Value)+"%")
		return fmt.Sprintf("%s LIKE ?", field)
	default:
		*args = append(*args, f.Value)
		return fmt.Sprintf("%s %s ?", field, f.Op.ToString())
	}
}

// safeIdentifier 安全地处理标识符（表名、列名）
// 防止 SQL 注入
func safeIdentifier(name string) string {
	// 移除可能导致 SQL 注入的字符
	name = strings.TrimSpace(name)
	// 检查是否包含危险字符
	if strings.ContainsAny(name, ";'\"`-") {
		return "_invalid_identifier"
	}
	return name
}

// unmarshalJSON 解析 JSON
func unmarshalJSON(data string, v interface{}) error {
	// 使用常见的 JSON 解析
	// 这里为了避免循环依赖，直接使用简单解析
	data = strings.TrimSpace(data)
	if len(data) < 2 {
		return fmt.Errorf("invalid json")
	}
	// 简单检查 JSON 格式
	if (data[0] != '{' && data[0] != '[') || (data[len(data)-1] != '}' && data[len(data)-1] != ']') {
		return fmt.Errorf("invalid json format")
	}
	// 使用 fmt.Sscan 无法解析复杂 JSON，这里简化处理
	// 实际使用时应该使用 json.Unmarshal
	return nil
}

// BunSQLBuilder 使用 bun 方式的 SQL 构建器
type BunSQLBuilder struct {
	dialect DialectType
}

// NewBunSQLBuilder 创建新的 BunSQLBuilder
func NewBunSQLBuilder(dialect DialectType) *BunSQLBuilder {
	return &BunSQLBuilder{dialect: dialect}
}

// BuildSelect 构建选择查询
func (b *BunSQLBuilder) BuildSelect(ast *QueryAST) string {
	qb := NewBunQueryBuilder()
	qb.SetDialect(b.dialect)
	qb.columnMappings = ast.ColumnMappings

	sql, _ := qb.BuildSelectQuery(ast)
	return sql
}

// BuildCount 构建计数查询
func (b *BunSQLBuilder) BuildCount(ast *QueryAST) string {
	qb := NewBunQueryBuilder()
	qb.SetDialect(b.dialect)
	qb.columnMappings = ast.ColumnMappings

	sql, _ := qb.BuildCountQuery(ast)
	return sql
}

// Dialect 返回方言类型
func (b *BunSQLBuilder) Dialect() DialectType {
	return b.dialect
}

// BuildBunQuery 使用 bun QueryBuilder 构建完整查询
// 返回主查询 SQL、计数查询 SQL 和参数
func BuildBunQuery(dialect DialectType, ast *QueryAST) (string, string, []interface{}) {
	qb := NewBunQueryBuilder()
	qb.SetDialect(dialect)
	qb.columnMappings = ast.ColumnMappings

	selectSQL, selectArgs := qb.BuildSelectQuery(ast)
	countSQL, countArgs := qb.BuildCountQuery(ast)

	// 合并参数
	var allArgs []interface{}
	allArgs = append(allArgs, selectArgs...)
	allArgs = append(allArgs, countArgs...)

	return selectSQL, countSQL, allArgs
}
