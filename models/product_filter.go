package models

import "github.com/shopspring/decimal"

// ProductListFilter holds query parameters for listing products.
type ProductListFilter struct {
	Offset        int
	Limit         int
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}
