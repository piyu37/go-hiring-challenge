package models

import "github.com/shopspring/decimal"

// EffectiveVariantPrice returns the variant price when set, otherwise the product base price.
func EffectiveVariantPrice(variantPrice *decimal.Decimal, productPrice decimal.Decimal) decimal.Decimal {
	if variantPrice != nil {
		return *variantPrice
	}
	return productPrice
}
