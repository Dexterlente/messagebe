package util

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit       int `json:"limit"`
	Offset      int `json:"offset"`
	CurrentPage int `json:"current_page"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

func GetPagination(r *http.Request) Pagination {
	const defaultLimit = 10
	const defaultPage = 1

	limit := defaultLimit
	page := defaultPage

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	offset := (page - 1) * limit

	return Pagination{
		Limit:       limit,
		Offset:      offset,
		CurrentPage: page,
	}
}
