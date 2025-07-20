package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/service/product"
)

func SetupRoutesProducts(r *gin.RouterGroup, DB *sql.DB) {
	productRepo := product.NewProductRepository(DB)
	productService := product.NewProductService(productRepo)
	productHandler := product.NewProductHandler(productService)

	protected := r.Group("/products")
	{
		protected.GET("", productHandler.GetProducts)
	}
}
