package method

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/types"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type MethodHandler struct {
	methodService *Service
}

func NewMethodHandler(service *Service) *MethodHandler {
	return &MethodHandler{
		methodService: service,
	}
}

func (handler *MethodHandler) Create(c *gin.Context) {
	var input types.CreateMethodData
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	data, err := handler.methodService.Create(c.Request.Context(), input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create method", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Method created successfully", data)
}

func (h *MethodHandler) GetAll(c *gin.Context) {
	// Menggunakan pagination utility
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.DefaultQuery("search", "")
	filterType := c.DefaultQuery("type", "")
	active := c.DefaultQuery("status", "")

	// Hitung pagination
	paginationResult := utils.CalculatePagination(&page, &limit)

	// Panggil service dengan utils parameters
	data, totalCount, err := h.methodService.GetAll(
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

func (h *MethodHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID parameter", err.Error())
		return
	}

	var input types.UpdateMethodData
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	update, err := h.methodService.Update(c.Request.Context(), id, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update Method", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Method updated successfully", update)
}

func (h *MethodHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID parameter", err.Error())
		return
	}

	err = h.methodService.Delete(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete Method", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Method deleted successfully", nil)
}
