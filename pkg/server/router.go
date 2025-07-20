package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/handler"
	"github.com/wafi04/backendvazzz/service/category"
)

func SetupRoutes(r *gin.RouterGroup, db *sql.DB) {
	categoryRepo := category.NewCategoryRepository(db)
	categoryService := category.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	categoryGroup := r.Group("/categories")
	{
		categoryGroup.POST("", categoryHandler.CreateCategory)
		categoryGroup.GET("", categoryHandler.GetAllCategories)
		categoryGroup.GET("/:code", categoryHandler.GetCategoryByCode)
		categoryGroup.PUT("/:id", categoryHandler.UpdateCategory)
		categoryGroup.DELETE("/:id", categoryHandler.DeleteCategory)
	}
}
