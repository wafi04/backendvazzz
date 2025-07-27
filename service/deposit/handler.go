package deposit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type DepositHandler struct {
	service *DepositService
}

func NewDepositHandler(service *DepositService) *DepositHandler {
	return &DepositHandler{
		service: service,
	}
}

type CreateDepositRequest struct {
	Amount     int    `json:"amount" binding:"required,min=1"`
	MethodCode string `json:"method" binding:"required"`
}

func (h *DepositHandler) Create(c *gin.Context) {
	var req CreateDepositRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error(), err.Error())
		return
	}

	usernameInterface, exists := c.Get("username")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "username not found in context") // Fix: status code
		return
	}

	username, ok := usernameInterface.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "invalid username type") // Fix: status code
		return
	}

	depositID, err := h.service.Create(c.Request.Context(), req.Amount, req.MethodCode, username)
	if depositID == "" || err != nil {

		errorMsg := "Failed to create deposit"
		if err != nil {
			errorMsg = err.Error()
		}

		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create deposit", errorMsg) // Fix: status code & message
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "created deposit successfully", gin.H{
		"depositId": depositID,
	})
}
func (h *DepositHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.DefaultQuery("search", "")
	status := c.DefaultQuery("status", "")

	paginationResult := utils.CalculatePagination(&page, &limit)
	data, totalCount, err := h.service.GetAll(
		c.Request.Context(),
		paginationResult.Skip,
		paginationResult.Take,
		search,
		status,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch deposits", err.Error())
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

func (h *DepositHandler) GetAllByUsername(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	usernameInterface, exists := c.Get("username")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "username not found in context") // Fix: status code
		return
	}

	username, ok := usernameInterface.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "invalid username type") // Fix: status code
		return
	}
	paginationResult := utils.CalculatePagination(&page, &limit)
	data, totalCount, err := h.service.GetAllByUsername(
		c.Request.Context(),
		paginationResult.Take,
		paginationResult.Skip,
		username,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch deposits", err.Error())
		return
	}

	// Buat paginated response
	response := utils.CreatePaginatedResponse(
		data,
		paginationResult.CurrentPage,
		paginationResult.ItemsPerPage,
		totalCount,
	)

	utils.SuccessResponse(c, http.StatusOK, "Deposits retrieved successfully", response)
}

func (h *DepositHandler) GetByDepositID(c *gin.Context) {
	depositID := c.Param("id")

	if depositID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "deposit id not found", "")
		return
	}

	data, err := h.service.GetByDepositID(c, depositID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to get deposit", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Deposit Retreived Successfully", data)

}

func (h *DepositHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID parameter", err.Error())
		return
	}
	err = h.service.Delete(c, id)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to get deposit", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Deposit Deleted Successfully", nil)
}
