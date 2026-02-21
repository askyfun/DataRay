package model

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/uptrace/bun"
)

func CreateTables(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bi_datasource (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'postgresql',
			host VARCHAR(255) NOT NULL,
			port INTEGER NOT NULL,
			database_name VARCHAR(255) NOT NULL,
			username VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// bi_dataset
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bi_dataset (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			datasource_id INTEGER REFERENCES bi_datasource(id) ON DELETE CASCADE,
			table_name VARCHAR(255),
			query_sql TEXT,
			query_type VARCHAR(50) NOT NULL DEFAULT 'table',
			mode VARCHAR(50) NOT NULL DEFAULT 'direct',
			accelerate_config JSONB,
			description TEXT,
			tags JSONB DEFAULT '[]',
			refresh_strategy JSONB,
			preview_data JSONB,
			quality_rules JSONB DEFAULT '[]',
			columns JSONB DEFAULT '[]',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// bi_dataset_lineage
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bi_dataset_lineage (
			id SERIAL PRIMARY KEY,
			upstream_dataset_id INTEGER REFERENCES bi_dataset(id) ON DELETE CASCADE,
			downstream_dataset_id INTEGER REFERENCES bi_dataset(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// bi_chart
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bi_chart (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			dataset_id INTEGER REFERENCES bi_dataset(id) ON DELETE CASCADE,
			chart_type VARCHAR(50) NOT NULL,
			config JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// bi_share
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bi_share (
			id SERIAL PRIMARY KEY,
			token VARCHAR(255) NOT NULL UNIQUE,
			chart_id INTEGER REFERENCES bi_chart(id) ON DELETE CASCADE,
			password VARCHAR(255),
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_dataset_datasource_id ON bi_dataset(datasource_id);
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_chart_dataset_id ON bi_chart(dataset_id);
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_share_token ON bi_share(token);
	`); err != nil {
		return err
	}

	return nil
}

type Datasource struct {
	bun.BaseModel `bun:"bi_datasource"`

	ID           int          `bun:"id,pk,autoincrement" json:"id"`
	Name         string       `bun:"name" json:"name"`
	Type         string       `bun:"type" json:"type"`
	Host         string       `bun:"host" json:"host"`
	Port         int          `bun:"port" json:"port"`
	DatabaseName string       `bun:"database_name" json:"database_name"`
	Username     string       `bun:"username" json:"username"`
	Password     string       `bun:"password" json:"password"`
	CreatedAt    sql.NullTime `bun:"created_at" json:"created_at"`
	UpdatedAt    sql.NullTime `bun:"updated_at" json:"updated_at"`
}

// Dataset represents a data set
type Dataset struct {
	bun.BaseModel `bun:"bi_dataset"`

	ID               int            `bun:"id,pk,autoincrement" json:"id"`
	Name             string         `bun:"name" json:"name"`
	DatasourceID     int            `bun:"datasource_id" json:"datasource_id"`
	TableName        sql.NullString `bun:"table_name" json:"table_name"`
	QuerySQL         sql.NullString `bun:"query_sql" json:"query_sql"`
	QueryType        string         `bun:"query_type" json:"query_type"`
	Mode             string         `bun:"mode" json:"mode"`
	AccelerateConfig sql.NullString `bun:"accelerate_config" json:"accelerate_config"`
	Description      sql.NullString `bun:"description" json:"description"`
	Tags             string         `bun:"tags" json:"tags"`
	RefreshStrategy  sql.NullString `bun:"refresh_strategy" json:"refresh_strategy"`
	PreviewData      sql.NullString `bun:"preview_data" json:"preview_data"`
	QualityRules     string         `bun:"quality_rules" json:"quality_rules"`
	Columns          string         `bun:"columns" json:"columns"`
	CreatedAt        sql.NullTime   `bun:"created_at" json:"created_at"`
	UpdatedAt        sql.NullTime   `bun:"updated_at" json:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for Dataset
func (d Dataset) MarshalJSON() ([]byte, error) {
	type Alias Dataset
	aux := &struct {
		TableName        any `json:"table_name"`
		QuerySQL         any `json:"query_sql"`
		AccelerateConfig any `json:"accelerate_config"`
		Description      any `json:"description"`
		RefreshStrategy  any `json:"refresh_strategy"`
		PreviewData      any `json:"preview_data"`
		CreatedAt        any `json:"created_at"`
		UpdatedAt        any `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(&d),
	}

	if d.TableName.Valid {
		aux.TableName = d.TableName.String
	} else {
		aux.TableName = nil
	}
	if d.QuerySQL.Valid {
		aux.QuerySQL = d.QuerySQL.String
	} else {
		aux.QuerySQL = nil
	}
	if d.AccelerateConfig.Valid {
		aux.AccelerateConfig = d.AccelerateConfig.String
	} else {
		aux.AccelerateConfig = nil
	}
	if d.Description.Valid {
		aux.Description = d.Description.String
	} else {
		aux.Description = nil
	}
	if d.RefreshStrategy.Valid {
		aux.RefreshStrategy = d.RefreshStrategy.String
	} else {
		aux.RefreshStrategy = nil
	}
	if d.PreviewData.Valid {
		aux.PreviewData = d.PreviewData.String
	} else {
		aux.PreviewData = nil
	}
	if d.CreatedAt.Valid {
		aux.CreatedAt = d.CreatedAt.Time
	} else {
		aux.CreatedAt = nil
	}
	if d.UpdatedAt.Valid {
		aux.UpdatedAt = d.UpdatedAt.Time
	} else {
		aux.UpdatedAt = nil
	}

	return json.Marshal(aux)
}

type DatasetColumn struct {
	Name       string           `json:"name"`
	Expr       string           `json:"expr"`
	Type       StandardDataType `json:"type"`
	TypeConfig TypeConfig       `json:"type_config"`
	Comment    string           `json:"comment"`
	Role       string           `json:"role"`
}

// DatasetLineage represents lineage relationship between datasets
type DatasetLineage struct {
	bun.BaseModel `bun:"bi_dataset_lineage"`

	ID                  int          `bun:"id,pk,autoincrement" json:"id"`
	UpstreamDatasetID   int          `bun:"upstream_dataset_id" json:"upstream_dataset_id"`
	DownstreamDatasetID int          `bun:"downstream_dataset_id" json:"downstream_dataset_id"`
	CreatedAt           sql.NullTime `bun:"created_at" json:"created_at"`
}

// Chart represents a chart configuration
type Chart struct {
	bun.BaseModel `bun:"bi_chart"`

	ID        int          `bun:"id,pk,autoincrement" json:"id"`
	Name      string       `bun:"name" json:"name"`
	DatasetID int          `bun:"dataset_id" json:"dataset_id"`
	ChartType string       `bun:"chart_type" json:"chart_type"`
	Config    string       `bun:"config" json:"config"`
	CreatedAt sql.NullTime `bun:"created_at" json:"created_at"`
	UpdatedAt sql.NullTime `bun:"updated_at" json:"updated_at"`
}

// Share represents a shared chart
type Share struct {
	bun.BaseModel `bun:"bi_share"`

	ID        int            `bun:"id,pk,autoincrement" json:"id"`
	Token     string         `bun:"token" json:"token"`
	ChartID   int            `bun:"chart_id" json:"chart_id"`
	Password  sql.NullString `bun:"password" json:"password"`
	ExpiresAt sql.NullTime   `bun:"expires_at" json:"expires_at"`
	CreatedAt sql.NullTime   `bun:"created_at" json:"created_at"`
}
