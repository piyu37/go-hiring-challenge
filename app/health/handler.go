package health

import (
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandleLive(w http.ResponseWriter, _ *http.Request) {
	api.OKResponse(w, map[string]string{"status": "ok"})
}

func (h *Handler) HandleReady(w http.ResponseWriter, r *http.Request) {
	if err := database.Ping(r.Context(), h.db); err != nil {
		api.ErrorResponse(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	api.OKResponse(w, map[string]string{"status": "ready"})
}
