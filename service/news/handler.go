package news

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type NewsHandler struct {
	service *NewsService
}

func NewNewsHandler(service *NewsService) *NewsHandler {
	return &NewsHandler{service: service}
}

// ✅ POST /news
func (h *NewsHandler) Create(c *gin.Context) {
	var req model.CreateNews
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	news, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create news"})
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "News Created Successfully", news)
}

// ✅ GET /news?status=...&type=...
func (h *NewsHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	typ := c.Query("type")

	var statusPtr, typePtr *string
	if status != "" {
		statusPtr = &status
	}
	if typ != "" {
		typePtr = &typ
	}

	newsList, err := h.service.GetAll(statusPtr, typePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get news"})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News Retreived Successfully", newsList)
}

// ✅ GET /news/:id
func (h *NewsHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	news, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get news"})
		return
	}
	if news == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News Retreived Successfully", news)
}

// ✅ PUT /news/:id
func (h *NewsHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var req model.CreateNews
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update news"})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News Updated Successfully", updated)
}

// ✅ DELETE /news/:id
func (h *NewsHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete news"})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News delete Successfully", nil)
}
