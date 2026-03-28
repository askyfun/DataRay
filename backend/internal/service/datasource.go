package service

import (
	"context"
	"fmt"

	"dataray/internal/datasource"
	"dataray/internal/model"

	"github.com/uptrace/bun"
)

type DatasourceService struct {
	db *bun.DB
}

func NewDatasourceService(db *bun.DB) *DatasourceService {
	return &DatasourceService{db: db}
}

func (s *DatasourceService) GetByID(ctx context.Context, id int) (*model.Datasource, error) {
	ds := &model.Datasource{ID: id}
	if err := s.db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %v", err)
	}
	return ds, nil
}

func (s *DatasourceService) List(ctx context.Context, limit, offset int) ([]model.Datasource, error) {
	var datasources []model.Datasource
	query := s.db.NewSelect().Model(&datasources)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return datasources, nil
}

func (s *DatasourceService) Create(ctx context.Context, ds *model.Datasource) (*model.Datasource, error) {
	if _, err := s.db.NewInsert().Model(ds).Returning("*").Exec(ctx); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasourceService) Update(ctx context.Context, ds *model.Datasource) (*model.Datasource, error) {
	if _, err := s.db.NewUpdate().Model(ds).WherePK().Exec(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, ds.ID)
}

func (s *DatasourceService) Delete(ctx context.Context, id int) error {
	if _, err := s.db.NewDelete().Model(&model.Datasource{}).Where("id = ?", id).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *DatasourceService) GetConnection(ctx context.Context, datasourceID int) (datasource.Connection, error) {
	ds, err := s.GetByID(ctx, datasourceID)
	if err != nil {
		return nil, err
	}

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
	conn, err := driver.Connect(ctx, config)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return conn, nil
}

func (s *DatasourceService) TestConnection(ctx context.Context, driverType datasource.DriverType, config datasource.ConnectionConfig) error {
	driver, err := datasource.NewDriver(driverType)
	if err != nil {
		return fmt.Errorf("unsupported driver type: %s", driverType)
	}

	if err := driver.TestConnection(ctx, config); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}
