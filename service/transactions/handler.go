package transactions

import (
	"encoding/json"
	"fmt"
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

type RawCallbackRequest struct {
	Data         interface{} `json:"data,omitempty"`
	RefID        *string     `json:"ref_id,omitempty"`
	BuyerSKUCode *string     `json:"buyer_sku_code,omitempty"`
	CustomerNo   *string     `json:"customer_no,omitempty"`
	Status       *string     `json:"status,omitempty"`
	Message      *string     `json:"message,omitempty"`
	SN           *string     `json:"sn,omitempty"`
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

func (h *TransactionHandler) Invoice(c *gin.Context) {
	idParam := c.Param("id")

	data, err := h.transactionRepo.GetInvoiceByID(idParam)
	if err != nil {
		log.Printf("Repository error: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch transactions", err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", data)

}
func (h *TransactionHandler) CallbackDuitku(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	// Read raw request body
	rawBody, err := c.GetRawData()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}

	log.Printf("Raw callback body: %s", string(rawBody))

	// Process callback
	err = h.transactionRepo.CallbackTransactionFromDuitkuRaw(c, rawBody)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process callback", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Callback processed successfully", nil)
}

func (h *TransactionHandler) CallbackDigiflazz(c *gin.Context) {
	// Set response content type
	c.Header("Content-Type", "application/json")

	// Get raw request body for logging
	rawBody, exists := c.Get(gin.BodyBytesKey)
	if exists {
		log.Printf("Raw callback body: %s", string(rawBody.([]byte)))
	}

	var rawRequest RawCallbackRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&rawRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid json format", "Failed to unmarshal callback detail: %v")

		return
	}

	// Validate request body is not empty
	if rawRequest.Data == nil && rawRequest.RefID == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Empty is required", "Failed to unmarshal callback detail: %v")

		return
	}

	// Transform to standard callback data structure
	var callbackData CallbackData

	if rawRequest.Data != nil {
		dataBytes, err := json.Marshal(rawRequest.Data)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Failed to decode response from digiflazz", "Failed to unmarshal callback detail: %v")
			return
		}

		var detail CallbackDetail
		if err := json.Unmarshal(dataBytes, &detail); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Message is required", "Failed to unmarshal callback detail: %v")
			return
		}

		callbackData = CallbackData{
			Data: detail,
		}
	} else {
		callbackData = CallbackData{
			Data: CallbackDetail{
				RefID:        getStringValue(rawRequest.RefID),
				BuyerSKUCode: getStringValue(rawRequest.BuyerSKUCode),
				CustomerNo:   getStringValue(rawRequest.CustomerNo),
				Status:       getStringValue(rawRequest.Status),
				Message:      getStringValue(rawRequest.Message),
				SN:           getStringValue(rawRequest.SN),
			},
		}
	}

	// Validate required fields
	if callbackData.Data.RefID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "RefId is required", "RefId is required")
		return
	}

	if callbackData.Data.Status == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Message is required", "Message is required")
		return
	}

	// Process callback
	ctx := c.Request.Context()
	if err := h.transactionRepo.Callback(ctx, callbackData); err != nil {
		log.Printf("Callback processing failed: %v", err)

		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Processing failed: %v", err), err.Error())
		return
	}

	// Success response
	utils.SuccessResponse(c, http.StatusCreated, "Callback processed successfully", callbackData)
}

func (h *TransactionHandler) GetRepostTransaction(c *gin.Context) {
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
	data, totalCount, err := h.transactionRepo.GetReportTransactions(
		c.Request.Context(),
		paginationResult.ItemsPerPage,
		paginationResult.CurrentPage,
		username,
	)
	log.Println(data)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch report transactions", err.Error())
		return
	}

	// Buat paginated response
	response := utils.CreatePaginatedResponse(
		data,
		paginationResult.CurrentPage,
		paginationResult.ItemsPerPage,
		totalCount,
	)

	utils.SuccessResponse(c, http.StatusOK, "Report Trasactions retrieved successfully", response)
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
