package models

import "context"

// ProductStore abstracts product persistence operations used by HTTP handlers.
type ProductStore interface {
	List(ctx context.Context, filter ProductListFilter) ([]Product, int64, error)
	GetByCode(ctx context.Context, code string) (*Product, error)
}

// CategoryStore abstracts category persistence operations used by HTTP handlers.
type CategoryStore interface {
	ListAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, category *Category) error
	ExistsByCode(ctx context.Context, code string) (bool, error)
}
