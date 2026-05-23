package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProductStore struct {
	listFn      func(ctx context.Context, filter models.ProductListFilter) ([]models.Product, int64, error)
	getByCodeFn func(ctx context.Context, code string) (*models.Product, error)
}

func (m *mockProductStore) List(ctx context.Context, filter models.ProductListFilter) ([]models.Product, int64, error) {
	return m.listFn(ctx, filter)
}

func (m *mockProductStore) GetByCode(ctx context.Context, code string) (*models.Product, error) {
	return m.getByCodeFn(ctx, code)
}

func sampleProduct() models.Product {
	return models.Product{
		Code:  "PROD001",
		Price: decimal.NewFromFloat(10.99),
		Category: models.Category{
			Code: "clothing",
			Name: "Clothing",
		},
		Variants: []models.Variant{
			{Name: "Variant A", SKU: "SKU001A", Price: ptrDecimal(11.99)},
			{Name: "Variant B", SKU: "SKU001B", Price: nil},
		},
	}
}

func ptrDecimal(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}

func TestHandleList(t *testing.T) {
	t.Run("returns paginated products with defaults", func(t *testing.T) {
		store := &mockProductStore{
			listFn: func(_ context.Context, filter models.ProductListFilter) ([]models.Product, int64, error) {
				assert.Equal(t, 0, filter.Offset)
				assert.Equal(t, 10, filter.Limit)
				return []models.Product{sampleProduct()}, 1, nil
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body ListResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, int64(1), body.Total)
		require.Len(t, body.Products, 1)
		assert.Equal(t, "PROD001", body.Products[0].Code)
		assert.Equal(t, "clothing", body.Products[0].Category.Code)
	})

	t.Run("applies offset and limit query params", func(t *testing.T) {
		store := &mockProductStore{
			listFn: func(_ context.Context, filter models.ProductListFilter) ([]models.Product, int64, error) {
				assert.Equal(t, 5, filter.Offset)
				assert.Equal(t, 20, filter.Limit)
				return nil, 0, nil
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=5&limit=20", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("applies category and price_less_than filters", func(t *testing.T) {
		store := &mockProductStore{
			listFn: func(_ context.Context, filter models.ProductListFilter) ([]models.Product, int64, error) {
				assert.Equal(t, "clothing", filter.CategoryCode)
				require.NotNil(t, filter.PriceLessThan)
				assert.True(t, filter.PriceLessThan.Equal(decimal.NewFromFloat(20)))
				return nil, 0, nil
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=clothing&price_less_than=20", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns 400 for invalid limit", func(t *testing.T) {
		handler := NewHandler(&mockProductStore{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=0", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid price_less_than", func(t *testing.T) {
		handler := NewHandler(&mockProductStore{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?price_less_than=abc", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		store := &mockProductStore{
			listFn: func(context.Context, models.ProductListFilter) ([]models.Product, int64, error) {
				return nil, 0, errors.New("db down")
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestHandleGetByCode(t *testing.T) {
	t.Run("returns product details with inherited variant prices", func(t *testing.T) {
		product := sampleProduct()
		store := &mockProductStore{
			getByCodeFn: func(_ context.Context, code string) (*models.Product, error) {
				assert.Equal(t, "PROD001", code)
				return &product, nil
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body DetailResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, "PROD001", body.Code)
		assert.Equal(t, "Clothing", body.Category.Name)
		require.Len(t, body.Variants, 2)
		assert.Equal(t, 11.99, body.Variants[0].Price)
		assert.Equal(t, 10.99, body.Variants[1].Price)
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		store := &mockProductStore{
			getByCodeFn: func(context.Context, string) (*models.Product, error) {
				return nil, models.ErrNotFound
			},
		}

		handler := NewHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/catalog/UNKNOWN", nil)
		req.SetPathValue("code", "UNKNOWN")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestParseListParams(t *testing.T) {
	t.Run("defaults offset and limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		params, err := parseListParams(req)
		require.NoError(t, err)
		assert.Equal(t, 0, params.Offset)
		assert.Equal(t, 10, params.Limit)
	})

	t.Run("rejects limit above maximum", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=101", nil)
		_, err := parseListParams(req)
		assert.Error(t, err)
	})
}
