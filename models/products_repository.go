package models

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) scopedQuery(filter ProductListFilter) *gorm.DB {
	query := r.db.Model(&Product{})

	if filter.CategoryCode != "" {
		query = query.Where(
			"category_id = (SELECT id FROM categories WHERE code = ?)",
			filter.CategoryCode,
		)
	}

	if filter.PriceLessThan != nil {
		query = query.Where("price < ?", filter.PriceLessThan)
	}

	return query
}

func (r *ProductsRepository) List(ctx context.Context, filter ProductListFilter) ([]Product, int64, error) {
	query := r.scopedQuery(filter)

	var total int64
	if err := query.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []Product
	err := query.WithContext(ctx).
		Preload("Category").
		Order("id ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetByCode(ctx context.Context, code string) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants").
		Where("code = ?", code).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &product, nil
}
