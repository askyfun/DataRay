package model

import (
	"testing"
)

func TestStarRocksMapper_ToStandard(t *testing.T) {
	m := &StarRocksMapper{}

	tests := []struct {
		name       string
		sourceType string
		want       StandardDataType
		wantErr    bool
	}{
		// 数值类型
		{"int", "int", TypeInteger, false},
		{"bigint", "bigint", TypeInteger, false},
		{"tinyint", "tinyint", TypeInteger, false},
		{"smallint", "smallint", TypeInteger, false},
		{"largeint", "largeint", TypeInteger, false},
		{"float", "float", TypeNumber, false},
		{"double", "double", TypeNumber, false},
		{"decimal", "decimal", TypeNumber, false},
		{"decimal with params", "decimal(10,2)", TypeNumber, false},

		// 布尔类型
		{"bool", "bool", TypeBoolean, false},
		{"boolean", "boolean", TypeBoolean, false},

		// 字符串类型
		{"varchar", "varchar", TypeString, false},
		{"string", "string", TypeString, false},
		{"char", "char", TypeString, false},

		// 日期时间类型
		{"date", "date", TypeDate, false},
		{"datetime", "datetime", TypeDateTime, false},
		{"timestamp", "timestamp", TypeDateTime, false},

		// 复杂类型
		{"array", "array", TypeArray, false},
		{"map", "map", TypeMap, false},
		{"json", "json", TypeJSON, false},

		// 未知类型
		{"unknown", "unknown_type", TypeUnknown, true},
		{"empty", "", TypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := m.ToStandard(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStarRocksMapper_ToSource(t *testing.T) {
	m := &StarRocksMapper{}

	tests := []struct {
		name    string
		stdType StandardDataType
		config  TypeConfig
		want    string
	}{
		{"number default", TypeNumber, TypeConfig{}, "double"},
		{"number with precision", TypeNumber, TypeConfig{Precision: 10, Scale: 2}, "decimal(10,2)"},
		{"integer", TypeInteger, TypeConfig{}, "bigint"},
		{"boolean", TypeBoolean, TypeConfig{}, "boolean"},
		{"string", TypeString, TypeConfig{}, "varchar"},
		{"date", TypeDate, TypeConfig{}, "date"},
		{"datetime", TypeDateTime, TypeConfig{}, "datetime"},
		{"array", TypeArray, TypeConfig{}, "array"},
		{"map", TypeMap, TypeConfig{}, "map"},
		{"json", TypeJSON, TypeConfig{}, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ToSource(tt.stdType, tt.config)
			if got != tt.want {
				t.Errorf("ToSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLMapper_ToStandard(t *testing.T) {
	m := &PostgreSQLMapper{}

	tests := []struct {
		name       string
		sourceType string
		want       StandardDataType
		wantErr    bool
	}{
		// 数值类型
		{"integer", "integer", TypeInteger, false},
		{"smallint", "smallint", TypeInteger, false},
		{"bigint", "bigint", TypeInteger, false},
		{"real", "real", TypeNumber, false},
		{"double precision", "double precision", TypeNumber, false},
		{"numeric", "numeric", TypeNumber, false},
		{"decimal", "decimal", TypeNumber, false},
		{"numeric with params", "numeric(10,2)", TypeNumber, false},

		// 布尔类型
		{"boolean", "boolean", TypeBoolean, false},
		{"bool", "bool", TypeBoolean, false},

		// 字符串类型
		{"varchar", "varchar", TypeString, false},
		{"text", "text", TypeString, false},
		{"char", "char", TypeString, false},

		// 日期时间类型
		{"date", "date", TypeDate, false},
		{"timestamp", "timestamp", TypeDateTime, false},
		{"timestamptz", "timestamptz", TypeDateTime, false},

		// 复杂类型
		{"json", "json", TypeJSON, false},
		{"jsonb", "jsonb", TypeJSON, false},

		// 未知类型
		{"unknown", "unknown", TypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := m.ToStandard(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLMapper_ToSource(t *testing.T) {
	m := &PostgreSQLMapper{}

	tests := []struct {
		name    string
		stdType StandardDataType
		config  TypeConfig
		want    string
	}{
		{"number default", TypeNumber, TypeConfig{}, "numeric"},
		{"number with precision", TypeNumber, TypeConfig{Precision: 10, Scale: 2}, "numeric(10,2)"},
		{"integer", TypeInteger, TypeConfig{}, "bigint"},
		{"boolean", TypeBoolean, TypeConfig{}, "boolean"},
		{"string", TypeString, TypeConfig{}, "varchar"},
		{"date", TypeDate, TypeConfig{}, "date"},
		{"datetime", TypeDateTime, TypeConfig{}, "timestamp"},
		{"json", TypeJSON, TypeConfig{}, "jsonb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ToSource(tt.stdType, tt.config)
			if got != tt.want {
				t.Errorf("ToSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySQLMapper_ToStandard(t *testing.T) {
	m := &MySQLMapper{}

	tests := []struct {
		name       string
		sourceType string
		want       StandardDataType
		wantErr    bool
	}{
		// 数值类型
		{"int", "int", TypeInteger, false},
		{"tinyint", "tinyint", TypeInteger, false},
		{"smallint", "smallint", TypeInteger, false},
		{"mediumint", "mediumint", TypeInteger, false},
		{"bigint", "bigint", TypeInteger, false},
		{"float", "float", TypeNumber, false},
		{"double", "double", TypeNumber, false},
		{"decimal", "decimal", TypeNumber, false},
		{"decimal with params", "decimal(10,2)", TypeNumber, false},

		// 布尔类型（特殊处理）
		{"bool", "bool", TypeBoolean, false},
		{"boolean", "boolean", TypeBoolean, false},
		{"tinyint(1)", "tinyint(1)", TypeBoolean, false},

		// 字符串类型
		{"varchar", "varchar", TypeString, false},
		{"char", "char", TypeString, false},
		{"text", "text", TypeString, false},
		{"enum", "enum('a','b')", TypeString, false},

		// 日期时间类型
		{"date", "date", TypeDate, false},
		{"datetime", "datetime", TypeDateTime, false},
		{"timestamp", "timestamp", TypeDateTime, false},

		// 复杂类型
		{"json", "json", TypeJSON, false},

		// 未知类型
		{"unknown", "unknown", TypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := m.ToStandard(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySQLMapper_ToSource(t *testing.T) {
	m := &MySQLMapper{}

	tests := []struct {
		name    string
		stdType StandardDataType
		config  TypeConfig
		want    string
	}{
		{"number default", TypeNumber, TypeConfig{}, "decimal"},
		{"number with precision", TypeNumber, TypeConfig{Precision: 10, Scale: 2}, "decimal(10,2)"},
		{"integer", TypeInteger, TypeConfig{}, "bigint"},
		{"boolean", TypeBoolean, TypeConfig{}, "tinyint(1)"},
		{"string", TypeString, TypeConfig{}, "varchar"},
		{"date", TypeDate, TypeConfig{}, "date"},
		{"datetime", TypeDateTime, TypeConfig{}, "datetime"},
		{"json", TypeJSON, TypeConfig{}, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ToSource(tt.stdType, tt.config)
			if got != tt.want {
				t.Errorf("ToSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClickHouseMapper_ToStandard(t *testing.T) {
	m := &ClickHouseMapper{}

	tests := []struct {
		name       string
		sourceType string
		want       StandardDataType
		wantErr    bool
	}{
		// 数值类型
		{"int8", "int8", TypeInteger, false},
		{"int16", "int16", TypeInteger, false},
		{"int32", "int32", TypeInteger, false},
		{"int64", "int64", TypeInteger, false},
		{"int256", "int256", TypeInteger, false},
		{"uint8", "uint8", TypeInteger, false},
		{"uint64", "uint64", TypeInteger, false},
		{"float32", "float32", TypeNumber, false},
		{"float64", "float64", TypeNumber, false},
		{"decimal", "decimal", TypeNumber, false},
		{"decimal with params", "decimal(10,2)", TypeNumber, false},

		// 布尔类型
		{"bool", "bool", TypeBoolean, false},

		// 字符串类型
		{"string", "string", TypeString, false},
		{"fixedstring", "FixedString(16)", TypeString, false},
		{"uuid", "UUID", TypeString, false},

		// 日期时间类型
		{"date", "date", TypeDate, false},
		{"date32", "date32", TypeDate, false},
		{"datetime", "datetime", TypeDateTime, false},
		{"datetime64", "DateTime64", TypeDateTime, false},

		// 复杂类型
		{"array", "Array(Int8)", TypeArray, false},
		{"map", "Map(String, Int32)", TypeMap, false},
		{"json", "JSON", TypeJSON, false},
		{"object json", "Object('json')", TypeJSON, false},

		// 未知类型
		{"unknown", "unknown", TypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := m.ToStandard(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClickHouseMapper_ToSource(t *testing.T) {
	m := &ClickHouseMapper{}

	tests := []struct {
		name    string
		stdType StandardDataType
		config  TypeConfig
		want    string
	}{
		{"number default", TypeNumber, TypeConfig{}, "float64"},
		{"number with precision", TypeNumber, TypeConfig{Precision: 10, Scale: 2}, "decimal(10,2)"},
		{"integer", TypeInteger, TypeConfig{}, "int64"},
		{"boolean", TypeBoolean, TypeConfig{}, "bool"},
		{"string", TypeString, TypeConfig{}, "string"},
		{"date", TypeDate, TypeConfig{}, "date"},
		{"datetime", TypeDateTime, TypeConfig{}, "datetime"},
		{"array", TypeArray, TypeConfig{}, "array"},
		{"map", TypeMap, TypeConfig{}, "map"},
		{"json", TypeJSON, TypeConfig{}, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ToSource(tt.stdType, tt.config)
			if got != tt.want {
				t.Errorf("ToSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDataTypeMapper(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantName   string
		wantErr    bool
	}{
		{"starrocks", "starrocks", "starrocks", false},
		{"postgresql", "postgresql", "postgresql", false},
		{"postgres", "postgres", "postgresql", false},
		{"mysql", "mysql", "mysql", false},
		{"clickhouse", "clickhouse", "clickhouse", false},
		{"unknown", "unknown", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper, err := NewDataTypeMapper(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDataTypeMapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mapper.GetSourceName() != tt.wantName {
				t.Errorf("GetSourceName() = %v, want %v", mapper.GetSourceName(), tt.wantName)
			}
		})
	}
}

func TestStandardDataType_Helpers(t *testing.T) {
	tests := []struct {
		name       string
		stdType    StandardDataType
		isNumeric  bool
		isDateTime bool
		isComplex  bool
	}{
		{"number", TypeNumber, true, false, false},
		{"integer", TypeInteger, true, false, false},
		{"boolean", TypeBoolean, false, false, false},
		{"string", TypeString, false, false, false},
		{"date", TypeDate, false, true, false},
		{"datetime", TypeDateTime, false, true, false},
		{"array", TypeArray, false, false, true},
		{"map", TypeMap, false, false, true},
		{"json", TypeJSON, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stdType.IsNumeric(); got != tt.isNumeric {
				t.Errorf("IsNumeric() = %v, want %v", got, tt.isNumeric)
			}
			if got := tt.stdType.IsDateTime(); got != tt.isDateTime {
				t.Errorf("IsDateTime() = %v, want %v", got, tt.isDateTime)
			}
			if got := tt.stdType.IsComplex(); got != tt.isComplex {
				t.Errorf("IsComplex() = %v, want %v", got, tt.isComplex)
			}
		})
	}
}

func TestInferExpressionResultType(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want StandardDataType
	}{
		{"SUM", "SUM(price)", TypeNumber},
		{"AVG", "AVG(amount)", TypeNumber},
		{"COUNT", "COUNT(*)", TypeInteger},
		{"MAX", "MAX(score)", TypeNumber},
		{"MIN", "MIN(value)", TypeNumber},
		{"CONCAT", "CONCAT(first_name, last_name)", TypeString},
		{"UPPER", "UPPER(name)", TypeString},
		{"LOWER", "LOWER(title)", TypeString},
		{"LENGTH", "LENGTH(text)", TypeInteger},
		{"YEAR", "YEAR(created_at)", TypeInteger},
		{"MONTH", "MONTH(birth_date)", TypeInteger},
		{"ROUND", "ROUND(price, 2)", TypeNumber},
		{"ABS", "ABS(value)", TypeNumber},
		{"arithmetic multiply", "price * quantity", TypeNumber},
		{"arithmetic add", "amount + tax", TypeNumber},
		{"arithmetic modulo", "count % 10", TypeInteger},
		{"string concat operator", "first_name || last_name", TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferExpressionResultType(tt.expr)
			if got != tt.want {
				t.Errorf("InferExpressionResultType(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestAllDataSources_RoundTrip(t *testing.T) {
	sources := []struct {
		name   string
		mapper DataTypeMapper
		types  []string
	}{
		{
			name:   "StarRocks",
			mapper: &StarRocksMapper{},
			types:  []string{"int", "bigint", "float", "double", "decimal", "varchar", "date", "datetime", "boolean", "array", "map", "json"},
		},
		{
			name:   "PostgreSQL",
			mapper: &PostgreSQLMapper{},
			types:  []string{"integer", "bigint", "numeric", "real", "varchar", "text", "date", "timestamp", "boolean", "json", "jsonb"},
		},
		{
			name:   "MySQL",
			mapper: &MySQLMapper{},
			types:  []string{"int", "bigint", "decimal", "float", "varchar", "text", "date", "datetime", "timestamp", "tinyint(1)", "json"},
		},
		{
			name:   "ClickHouse",
			mapper: &ClickHouseMapper{},
			types:  []string{"int64", "float64", "decimal", "string", "date", "datetime", "bool", "array", "map", "json"},
		},
	}

	for _, source := range sources {
		t.Run(source.name, func(t *testing.T) {
			for _, srcType := range source.types {
				stdType, _, err := source.mapper.ToStandard(srcType)
				if err != nil {
					t.Errorf("ToStandard(%q) failed: %v", srcType, err)
					continue
				}
				if stdType == TypeUnknown {
					t.Errorf("ToStandard(%q) returned unknown type", srcType)
					continue
				}

				// 反向转换
				backType := source.mapper.ToSource(stdType, TypeConfig{})
				if backType == "" {
					t.Errorf("ToSource() returned empty string for %v", stdType)
				}
			}
		})
	}
}
