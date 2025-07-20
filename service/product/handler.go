package product

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type ProductHandler struct {
	productService *ProductService
}

func NewProductHandler(productService *ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	role, ok := c.Get("role")
	if !ok {
		role = "MEMBER"
	}

	userRole, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user role format",
		})
		return
	}

	// Parse query parameters
	categoryIdStr := c.Query("categoryId")
	subCategoryIdStr := c.Query("subCategoryId")

	var categoryId, subCategoryId int
	var err error

	if categoryIdStr != "" {
		categoryId, err = strconv.Atoi(categoryIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid category_id format",
			})
			return
		}
	}

	if subCategoryIdStr != "" {
		subCategoryId, err = strconv.Atoi(subCategoryIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid sub_category_id format",
			})
			return
		}
	}

	// Get products from repository
	products, err := h.productService.GetAll(categoryId, subCategoryId, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch products",
			"error":   err.Error(),
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product Retreived Successfully", products)
}
