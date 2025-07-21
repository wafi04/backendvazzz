package transactions

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type TransactionHandler struct {
	transactionRepo *TransactionsRepository
}

func NewTransactionHandler(trepo *TransactionsRepository) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: trepo,
	}
}
func (h *TransactionHandler) GetAll(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	status := c.Query("status")
	filterType := c.Query("transactionType")

	// Convert to integers
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Debug log untuk check parameter
	log.Printf("Handler params - page: %d, limit: %d, type: %s, status: %s, startDate: %s, endDate: %s",
		page, limit, filterType, status, startDate, endDate)

	// Build filter request
	filterReq := model.FilterTransaction{
		Limit:     limit, // ✅ Benar
		Page:      page,  // ✅ Benar
		Type:      filterType,
		Search:    &search,
		StartDate: &startDate,
		EndDate:   &endDate,
		Status:    &status,
	}

	// Set pointer fields hanya jika tidak kosong
	if search != "" {
		filterReq.Search = &search
	}
	if startDate != "" {
		filterReq.StartDate = &startDate
	}
	if endDate != "" {
		filterReq.EndDate = &endDate
	}
	if status != "" {
		filterReq.Status = &status
	}

	// Call repository
	data, totalCount, err := h.transactionRepo.GetAllWithPayment(c.Request.Context(), filterReq)
	if err != nil {
		log.Printf("Repository error: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch transactions", err.Error())
		return
	}

	log.Printf("Query result - count: %d, total: %d", len(data), totalCount)

	// Create paginated response
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	response := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"currentPage":  page,
			"totalPages":   totalPages,
			"totalItems":   totalCount,
			"itemsPerPage": limit,
			"hasNextPage":  page < totalPages,
			"hasPrevPage":  page > 1,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", response)
}
