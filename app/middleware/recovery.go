package middleware

import (
	"log"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v path=%s", recovered, r.URL.Path)
				api.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
