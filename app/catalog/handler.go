package catalog

import (
	"errors"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Handler struct {
	products models.ProductStore
}

func NewHandler(products models.ProductStore) *Handler {
	return &Handler{products: products}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	params, err := parseListParams(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	priceLessThan, err := parsePriceLessThan(params.PriceLessThan)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid price_less_than")
		return
	}

	products, total, err := h.products.List(r.Context(), models.ProductListFilter{
		Offset:        params.Offset,
		Limit:         params.Limit,
		CategoryCode:  params.CategoryCode,
		PriceLessThan: priceLessThan,
	})
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	response := ListResponse{
		Products: make([]ProductDTO, len(products)),
		Total:    total,
	}
	for i, p := range products {
		response.Products[i] = toProductDTO(p)
	}

	api.OKResponse(w, response)
}

func (h *Handler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.products.GetByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch product")
		return
	}

	api.OKResponse(w, toDetailResponse(*product))
}
