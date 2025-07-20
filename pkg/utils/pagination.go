package utils

import (
	"math"
	"strconv"
)

type PaginationParams struct {
	Limit *string `json:"limit,omitempty" form:"limit" query:"limit"`
	Page  *string `json:"page,omitempty" form:"page" query:"page"`
}

type FilterCategories struct {
	PaginationParams
	Search *string `json:"search,omitempty" form:"search" query:"search"`
	Status *string `json:"status,omitempty" form:"status" query:"status"`
}

type PaginationMeta struct {
	CurrentPage  int  `json:"currentPage"`
	TotalPages   int  `json:"totalPages"`
	TotalItems   int  `json:"totalItems"`
	ItemsPerPage int  `json:"itemsPerPage"`
	HasNextPage  bool `json:"hasNextPage"`
	HasPrevPage  bool `json:"hasPrevPage"`
}

type PaginatedResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationResult struct {
	Skip         int `json:"skip"`
	Take         int `json:"take"`
	CurrentPage  int `json:"currentPage"`
	ItemsPerPage int `json:"itemsPerPage"`
}

type PaginationUtil struct{}

func NewPaginationUtil() *PaginationUtil {
	return &PaginationUtil{}
}

func (p *PaginationUtil) CalculatePagination(page, limit *string) PaginationResult {
	currentPage := 1
	itemsPerPage := 10

	if page != nil {
		if parsed, err := strconv.Atoi(*page); err == nil && parsed > 0 {
			currentPage = parsed
		}
	}

	if limit != nil {
		if parsed, err := strconv.Atoi(*limit); err == nil && parsed > 0 {
			itemsPerPage = int(math.Min(100, float64(parsed))) // Max 100 items per page
		}
	}

	skip := (currentPage - 1) * itemsPerPage

	return PaginationResult{
		Skip:         skip,
		Take:         itemsPerPage,
		CurrentPage:  currentPage,
		ItemsPerPage: itemsPerPage,
	}
}

func (p *PaginationUtil) CreatePaginationMeta(currentPage, itemsPerPage, totalItems int) PaginationMeta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(itemsPerPage)))

	return PaginationMeta{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
		HasNextPage:  currentPage < totalPages,
		HasPrevPage:  currentPage > 1,
	}
}
func (p *PaginationUtil) CreatePaginatedResponseGeneric(
	data interface{},
	currentPage, itemsPerPage, totalItems int,
) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
		"meta": p.CreatePaginationMeta(currentPage, itemsPerPage, totalItems),
	}
}

func CalculatePagination(page, limit *string) PaginationResult {
	util := NewPaginationUtil()
	return util.CalculatePagination(page, limit)
}

func CreatePaginationMeta(currentPage, itemsPerPage, totalItems int) PaginationMeta {
	util := NewPaginationUtil()
	return util.CreatePaginationMeta(currentPage, itemsPerPage, totalItems)
}

func CreatePaginatedResponse[T any](
	data []T,
	currentPage, itemsPerPage, totalItems int,
) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data: data,
		Meta: CreatePaginationMeta(currentPage, itemsPerPage, totalItems),
	}
}
