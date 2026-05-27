package categories

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ListResponse struct {
	Categories []CategoryDTO `json:"categories"`
}

type createRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Handler struct {
	categories models.CategoryStore
}

func NewHandler(categories models.CategoryStore) *Handler {
	return &Handler{categories: categories}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.ListAll(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	response := ListResponse{
		Categories: make([]CategoryDTO, len(categories)),
	}
	for i, c := range categories {
		response.Categories[i] = CategoryDTO{
			Code: c.Code,
			Name: c.Name,
		}
	}

	api.OKResponse(w, response)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.categories.Create(r.Context(), category); err != nil {
		if errors.Is(err, models.ErrDuplicate) {
			api.ErrorResponse(w, http.StatusConflict, "category already exists")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	api.CreatedResponse(w, CategoryDTO{
		Code: category.Code,
		Name: category.Name,
	})
}
