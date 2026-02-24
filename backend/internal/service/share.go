package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"dataray/internal/model"

	"github.com/uptrace/bun"
)

type ShareService struct {
	db *bun.DB
}

func NewShareService(db *bun.DB) *ShareService {
	return &ShareService{db: db}
}

func (s *ShareService) GetByToken(ctx context.Context, token string) (*model.Share, error) {
	share := &model.Share{}
	if err := s.db.NewSelect().Model(share).Where("token = ?", token).Scan(ctx); err != nil {
		return nil, fmt.Errorf("share not found")
	}
	return share, nil
}

func (s *ShareService) Create(ctx context.Context, chartID int) (*model.Share, error) {
	token := generateToken()
	share := &model.Share{
		Token:   token,
		ChartID: chartID,
	}

	if _, err := s.db.NewInsert().Model(share).Exec(ctx); err != nil {
		return nil, err
	}

	return share, nil
}

func (s *ShareService) IsExpired(share *model.Share) bool {
	if share.ExpiresAt.Valid && share.ExpiresAt.Time.Before(time.Now()) {
		return true
	}
	return false
}

func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		now := time.Now().UnixNano()
		return fmt.Sprintf("%d-%d", now, now%100000000)
	}
	return fmt.Sprintf("%s-%x", hex.EncodeToString(b[:8]), b[8:])
}
