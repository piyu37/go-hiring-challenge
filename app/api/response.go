package api

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error string `json:"error"`
}

func OKResponse(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

func CreatedResponse(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
