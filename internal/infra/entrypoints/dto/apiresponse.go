package dto

import (
	"fmt"
	"net/url"
	"strings"
)

type (
	ApiResponseStatus string
)

var (
	DefaultPageSize                      = 25
	ApiResponseSuccess ApiResponseStatus = "success"
	ApiResponseFailed  ApiResponseStatus = "failed"
	QueryQueryParam                      = "query"
	SortQueryParam                       = "sort"
)

type SortOrder struct {
	Property  string
	Direction string
}

type Pageable interface {
	GetSort() []SortOrder
}

type Page[T any] interface {
	HasContent() bool
	GetNumber() int
	HasPrevious() bool
	HasNext() bool
	GetTotalPages() int
	GetNumberOfElements() int
	GetTotalElements() int64
	GetSize() int
	GetContent() []T
	GetSort() []SortOrder
	GetPageable() Pageable
}

type Pagination struct {
	CurrentPage  int    `json:"current_page,omitempty"`
	TotalItems   int64  `json:"total_items"`
	PageSize     int    `json:"page_size"`
	NextUri      string `json:"next_uri,omitempty"`
	PreviousUri  string `json:"previous_uri,omitempty"`
	NextPage     int    `json:"next_page,omitempty"`
	PreviousPage int    `json:"previous_page,omitempty"`
	TotalPage    int    `json:"total_page"`
}

type ApiResponse[T any] struct {
	Data   T                 `json:"data,omitempty"`
	Page   *Pagination       `json:"page,omitempty"`
	Status ApiResponseStatus `json:"status"`
}

type ErrorField struct {
	Message string `json:"message"`
}

// EmptyPage the empty mutable Pagination instance
func EmptyPage() *Pagination {
	return &Pagination{
		CurrentPage:  0,
		NextPage:     0,
		PreviousPage: 0,
		TotalItems:   0,
		TotalPage:    0,
		PageSize:     DefaultPageSize,
	}
}

func NilPage() *Pagination {
	return nil
}

func OfPage(total, current, prev, next int) *Pagination {
	return &Pagination{
		CurrentPage:  current,
		PageSize:     DefaultPageSize,
		NextPage:     next,
		PreviousPage: prev,
		TotalPage:    total,
	}
}

// ApiResponse function

// OkResponse the response when request succeeded
// dto <T> the response type
func OkResponse[T any](dto T) ApiResponse[T] {
	return ApiResponse[T]{
		Data:   dto,
		Page:   nil,
		Status: ApiResponseSuccess,
	}
} // FailResponse the response when request succeeded
func FailResponse[T any](dto T) ApiResponse[T] {
	return ApiResponse[T]{
		Data:   dto,
		Page:   nil,
		Status: ApiResponseFailed,
	}
}

func OkListResponse[T any](dtos ...T) ApiResponse[[]T] {
	return ApiResponse[[]T]{
		Data:   dtos,
		Page:   EmptyPage(),
		Status: ApiResponseSuccess,
	}
}

func EmptyResponse() ApiResponse[any] {
	return ApiResponse[any]{Status: ApiResponseSuccess}
}

// OkPage is the Go equivalent of the ApiResponse.of(Page<T>, ...) overloads.
//
// opts:
//
//	opts[0] = uri
//	opts[1] = query
//	opts[2] = sortBy
//
// Examples:
//
//	OkPage(page)
//	OkPage(page, uri)
//	OkPage(page, uri, query)
//	OkPage(page, uri, query, sortBy)
//
// go::deprecated
func OkPage[T any](page Page[T], opts ...string) ApiResponse[[]T] {
	var uri, query, sortBy string

	if len(opts) > 0 {
		uri = opts[0]
	}
	if len(opts) > 1 {
		query = opts[1]
	}
	if len(opts) > 2 {
		sortBy = opts[2]
	}

	if page == nil || !page.HasContent() {
		return ApiResponse[[]T]{
			Status: ApiResponseSuccess,
		}
	}

	number := page.GetNumber()

	prev := 0
	if page.HasPrevious() {
		prev = number - 1
	}

	next := 0
	if page.HasNext() {
		next = number + 1
	}

	pagination := &Pagination{
		CurrentPage:  number,
		TotalItems:   page.GetTotalElements(),
		PageSize:     page.GetNumberOfElements(),
		NextPage:     next,
		PreviousPage: prev,
		TotalPage:    page.GetTotalPages(),
	}

	if uri != "" {
		if sortBy == "" {
			sortBy = getSortFromPage(page)
		}

		sep := "?"
		if strings.Contains(uri, "?") {
			sep = "&"
		}

		uri = uri + sep +
			QueryQueryParam + "=" + encodeURL(query) +
			"&" + SortQueryParam + "=" + sortBy

		if page.HasNext() {
			pagination.NextUri = fmt.Sprintf("%s&page=%d&size=%d", uri, next, page.GetSize())
		}

		if page.HasPrevious() {
			pagination.PreviousUri = fmt.Sprintf("%s&page=%d&size=%d", uri, prev, page.GetSize())
		}
	}

	return ApiResponse[[]T]{
		Data:   page.GetContent(),
		Page:   pagination,
		Status: ApiResponseSuccess,
	}
}

func getSortFromPage[T any](page Page[T]) string {
	sort := page.GetSort()

	if len(sort) == 0 {
		if pageable := page.GetPageable(); pageable != nil {
			sort = pageable.GetSort()
		}
	}

	if len(sort) == 0 {
		return ""
	}

	parts := make([]string, 0, len(sort))
	for _, order := range sort {
		parts = append(parts, order.Property+","+order.Direction)
	}

	return strings.Join(parts, "&")
}

func encodeURL(value string) string {
	return url.QueryEscape(value)
}
