package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/handler"
	"github.com/wafi04/backendvazzz/service/subcategory"
)

func SetupRoutesSubCategories(r *gin.RouterGroup, DB *sql.DB) {
	subCategoryRepo := subcategory.NewSubCategory(DB)
	subCategoryService := subcategory.NewSubCategoryService(subCategoryRepo)
	subCategoryHandler := handler.NewSubCategoryHandler(subCategoryService)

	categoryGroup := r.Group("/subcategories")
	{
		categoryGroup.POST("", subCategoryHandler.CreateSubCategory)
		categoryGroup.GET("", subCategoryHandler.GetAllSubCategories)
		categoryGroup.GET("/:id", subCategoryHandler.GetSubCategoryByID)
		categoryGroup.PUT("/:id", subCategoryHandler.UpdateSubCategory)
		categoryGroup.DELETE("/:id", subCategoryHandler.DeleteSubCategory)
	}
}
