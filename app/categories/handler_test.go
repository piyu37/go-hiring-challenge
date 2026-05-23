package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCategoryStore struct {
	listAllFn func(ctx context.Context) ([]models.Category, error)
	createFn  func(ctx context.Context, category *models.Category) error
}

func (m *mockCategoryStore) ListAll(ctx context.Context) ([]models.Category, error) {
	if m.listAllFn == nil {
		return nil, nil
	}
	return m.listAllFn(ctx)
}

func (m *mockCategoryStore) Create(ctx context.Context, category *models.Category) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, category)
}

func (m *mockCategoryStore) ExistsByCode(context.Context, string) (bool, error) {
	return true, nil
}

func TestHandleList(t *testing.T) {
	store := &mockCategoryStore{
		listAllFn: func(context.Context) ([]models.Category, error) {
			return []models.Category{
				{Code: "clothing", Name: "Clothing"},
				{Code: "shoes", Name: "Shoes"},
			}, nil
		},
	}

	handler := NewHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()

	handler.HandleList(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body ListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Categories, 2)
	assert.Equal(t, "clothing", body.Categories[0].Code)
}

func TestHandleCreate(t *testing.T) {
	t.Run("creates category successfully", func(t *testing.T) {
		store := &mockCategoryStore{
			createFn: func(_ context.Context, category *models.Category) error {
				assert.Equal(t, "bags", category.Code)
				assert.Equal(t, "Bags", category.Name)
				return nil
			},
		}

		handler := NewHandler(store)
		body := bytes.NewBufferString(`{"code":"bags","name":"Bags"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response CategoryDTO
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "bags", response.Code)
		assert.Equal(t, "Bags", response.Name)
	})

	t.Run("returns 400 for missing fields", func(t *testing.T) {
		handler := NewHandler(&mockCategoryStore{})
		body := bytes.NewBufferString(`{"code":"","name":""}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 409 for duplicate category", func(t *testing.T) {
		store := &mockCategoryStore{
			createFn: func(context.Context, *models.Category) error {
				return models.ErrDuplicate
			},
		}

		handler := NewHandler(store)
		body := bytes.NewBufferString(`{"code":"clothing","name":"Clothing"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		store := &mockCategoryStore{
			createFn: func(context.Context, *models.Category) error {
				return errors.New("db down")
			},
		}

		handler := NewHandler(store)
		body := bytes.NewBufferString(`{"code":"bags","name":"Bags"}`)
		req := httptest.NewRequest(http.MethodPost, "/categories", body)
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
