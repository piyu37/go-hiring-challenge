package catalog

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	defaultOffset = 0
	defaultLimit  = 10
	minLimit      = 1
	maxLimit      = 100
)

type listParams struct {
	Offset        int
	Limit         int
	CategoryCode  string
	PriceLessThan string
}

func parseListParams(r *http.Request) (listParams, error) {
	offset := defaultOffset
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return listParams{}, fmt.Errorf("invalid offset")
		}
		offset = parsed
	}

	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < minLimit || parsed > maxLimit {
			return listParams{}, fmt.Errorf("invalid limit")
		}
		limit = parsed
	}

	return listParams{
		Offset:        offset,
		Limit:         limit,
		CategoryCode:  r.URL.Query().Get("category"),
		PriceLessThan: r.URL.Query().Get("price_less_than"),
	}, nil
}
