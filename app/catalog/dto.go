package catalog

import (
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type CategoryDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductDTO struct {
	Code     string      `json:"code"`
	Price    float64     `json:"price"`
	Category CategoryDTO `json:"category"`
}

type VariantDTO struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type ListResponse struct {
	Products []ProductDTO `json:"products"`
	Total    int64        `json:"total"`
}

type DetailResponse struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category CategoryDTO  `json:"category"`
	Variants []VariantDTO `json:"variants"`
}

func toCategoryDTO(c models.Category) CategoryDTO {
	return CategoryDTO{
		Code: c.Code,
		Name: c.Name,
	}
}

func toProductDTO(p models.Product) ProductDTO {
	return ProductDTO{
		Code:     p.Code,
		Price:    p.Price.InexactFloat64(),
		Category: toCategoryDTO(p.Category),
	}
}

func toDetailResponse(p models.Product) DetailResponse {
	variants := make([]VariantDTO, len(p.Variants))
	for i, v := range p.Variants {
		effective := models.EffectiveVariantPrice(v.Price, p.Price)
		variants[i] = VariantDTO{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: effective.InexactFloat64(),
		}
	}

	return DetailResponse{
		Code:     p.Code,
		Price:    p.Price.InexactFloat64(),
		Category: toCategoryDTO(p.Category),
		Variants: variants,
	}
}

func parsePriceLessThan(raw string) (*decimal.Decimal, error) {
	if raw == "" {
		return nil, nil
	}

	price, err := decimal.NewFromString(raw)
	if err != nil {
		return nil, err
	}

	return &price, nil
}
