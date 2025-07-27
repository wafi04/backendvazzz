package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *AnalyticsService
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewAnalyticsHandler(service *AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error:   "Analytics Error",
		Message: message,
	})
}

func (h *AnalyticsHandler) successResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// GET /analytics/range?start_date=2024-01-01&end_date=2024-01-31
func (h *AnalyticsHandler) GetAnalyticsByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		h.errorResponse(c, http.StatusBadRequest, "start_date and end_date parameters are required")
		return
	}

	results, err := h.service.GetAnalyticsByDateRange(startDate, endDate)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(c, "Analytics data retrieved successfully", results)
}

// GET /analytics/date?date=2024-01-15
func (h *AnalyticsHandler) GetAnalyticsByDate(c *gin.Context) {
	date := c.Query("date")

	if date == "" {
		h.errorResponse(c, http.StatusBadRequest, "date parameter is required")
		return
	}

	result, err := h.service.GetAnalyticsByDate(date)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(c, "Analytics data retrieved successfully", result)
}

// GET /analytics/today
func (h *AnalyticsHandler) GetTodayAnalytics(c *gin.Context) {
	result, err := h.service.GetTodayAnalytics()
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(c, "Today's analytics retrieved successfully", result)
}

// GET /analytics/monthly?year=2024&month=1
func (h *AnalyticsHandler) GetMonthlyAnalytics(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if yearStr == "" || monthStr == "" {
		// Default to current year and month
		now := time.Now()
		yearStr = strconv.Itoa(now.Year())
		monthStr = strconv.Itoa(int(now.Month()))
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid year parameter")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid month parameter")
		return
	}

	results, err := h.service.GetMonthlyAnalytics(year, month)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(c, "Monthly analytics retrieved successfully", results)
}

// GET /analytics/summary?start_date=2024-01-01&end_date=2024-01-31
func (h *AnalyticsHandler) GetAnalyticsSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		h.errorResponse(c, http.StatusBadRequest, "start_date and end_date parameters are required")
		return
	}

	result, err := h.service.GetAnalyticsSummary(startDate, endDate)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(c, "Analytics summary retrieved successfully", result)
}

// GET /analytics/status?start_date=2024-01-01&end_date=2024-01-31&status=success
func (h *AnalyticsHandler) GetAnalyticsByStatus(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	if startDate == "" || endDate == "" || status == "" {
		h.errorResponse(c, http.StatusBadRequest, "start_date, end_date, and status parameters are required")
		return
	}

	result, err := h.service.GetAnalyticsByStatus(startDate, endDate, status)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(c, "Analytics by status retrieved successfully", result)
}
