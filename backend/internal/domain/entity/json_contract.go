package entity

import "encoding/json"

// normalizeSlice converts nil slices to empty JSON arrays while preserving non-nil slices.
// It is used by API response structs whose collection fields must keep a stable array contract.
func normalizeSlice[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

// MarshalJSON serializes preview responses with empty arrays instead of null so frontend code
// can safely consume `columns` and `data` without special null handling.
func (r PreviewResult) MarshalJSON() ([]byte, error) {
	type previewResultAlias PreviewResult

	return json.Marshal(previewResultAlias{
		Columns: normalizeSlice(r.Columns),
		Data:    normalizeSlice(r.Data),
	})
}

// MarshalJSON serializes table data responses with empty arrays instead of null for all
// collection fields, keeping the pagination payload shape stable across empty results.
func (r TableDataResult) MarshalJSON() ([]byte, error) {
	type tableDataResultAlias TableDataResult

	return json.Marshal(tableDataResultAlias{
		Columns:     normalizeSlice(r.Columns),
		Data:        normalizeSlice(r.Data),
		Total:       r.Total,
		PrimaryKeys: normalizeSlice(r.PrimaryKeys),
		Page:        r.Page,
		PageSize:    r.PageSize,
	})
}

// MarshalJSON serializes field distribution responses with an empty distribution array instead
// of null, matching the API contract for collection-shaped response fields.
func (r FieldDistribution) MarshalJSON() ([]byte, error) {
	type fieldDistributionAlias FieldDistribution

	return json.Marshal(fieldDistributionAlias{
		FieldName:    r.FieldName,
		TotalCount:   r.TotalCount,
		UniqueCount:  r.UniqueCount,
		Distribution: normalizeSlice(r.Distribution),
	})
}
