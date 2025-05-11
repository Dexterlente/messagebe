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
	const defaultOffset = 0

	limit := defaultLimit
	offset := defaultOffset

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	currentPage := (offset / limit) + 1

	return Pagination{
		Limit:       limit,
		Offset:      offset,
		CurrentPage: currentPage,
	}
}
