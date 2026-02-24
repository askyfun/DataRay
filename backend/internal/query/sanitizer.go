package query

import (
	"fmt"
	"strconv"
	"strings"
)

func EscapeValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return escapeString(val)
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(toInt64(val), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(toUint64(val), 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return escapeString(fmt.Sprintf("%v", val))
	}
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	}
	return 0
}

func toUint64(v interface{}) uint64 {
	switch val := v.(type) {
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	}
	return 0
}
