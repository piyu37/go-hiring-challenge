package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestEffectiveVariantPrice(t *testing.T) {
	productPrice := decimal.NewFromFloat(10.99)
	variantPrice := decimal.NewFromFloat(11.99)

	t.Run("uses variant price when set", func(t *testing.T) {
		got := EffectiveVariantPrice(&variantPrice, productPrice)
		assert.True(t, got.Equal(variantPrice))
	})

	t.Run("inherits product price when variant price is nil", func(t *testing.T) {
		got := EffectiveVariantPrice(nil, productPrice)
		assert.True(t, got.Equal(productPrice))
	})
}
