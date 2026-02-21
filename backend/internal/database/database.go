package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"dataray/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func InitDB(databaseUrl string) (*bun.DB, error) {
	sqldb, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	slog.Info("Database connected successfully")
	return db, nil
}

func RunMigrations(db *bun.DB) error {
	ctx := context.Background()
	if err := model.CreateTables(ctx, db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := addDatasourceTypeColumn(ctx, db); err != nil {
		return fmt.Errorf("failed to run schema migrations: %w", err)
	}

	slog.Info("Migrations completed successfully")
	return nil
}

func addDatasourceTypeColumn(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		ALTER TABLE bi_datasource 
		ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT 'postgresql';
	`)
	return err
}
