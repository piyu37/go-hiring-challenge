package models

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) ListAll(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := r.db.WithContext(ctx).Order("id ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoriesRepository) Create(ctx context.Context, category *Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrDuplicate
	}

	return err
}

var ErrDuplicate = errors.New("duplicate record")
