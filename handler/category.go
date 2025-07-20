package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
	"github.com/wafi04/backendvazzz/service/category"
)

// Response struktur untuk standardize response

type CategoryHandler struct {
	categoryService *category.CategoryService
}

func NewCategoryHandler(categoryService *category.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var input model.CreateCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	err := h.categoryService.CreateCategory(c.Request.Context(), input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Category created successfully", nil)
}

func (h *CategoryHandler) GetCategoryByCode(c *gin.Context) {
	codeParam := c.Param("code")

	cat, err := h.categoryService.GetCategoryByCode(c.Request.Context(), codeParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", cat)
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	// Menggunakan pagination utility
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.DefaultQuery("search", "")
	filterType := c.DefaultQuery("type", "")
	active := c.DefaultQuery("status", "")

	// Hitung pagination
	paginationResult := utils.CalculatePagination(&page, &limit)

	// Panggil service dengan utils parameters
	data, totalCount, err := h.categoryService.GetAllCategories(
		c.Request.Context(),
		paginationResult.Skip,
		paginationResult.Take,
		search,
		filterType,
		active,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch categories", err.Error())
		return
	}

	// Buat paginated response
	response := utils.CreatePaginatedResponse(
		data,
		paginationResult.CurrentPage,
		paginationResult.ItemsPerPage,
		totalCount,
	)

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", response)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID parameter", err.Error())
		return
	}

	var input model.CreateCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	err = h.categoryService.UpdateCategory(c.Request.Context(), id, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category updated successfully", nil)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID parameter", err.Error())
		return
	}

	err = h.categoryService.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category deleted successfully", nil)
}
