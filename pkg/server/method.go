package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/service/method"
)

func SetupRoutesMethod(r *gin.RouterGroup, DB *sql.DB) {
	methodRepo := method.NewMethodRepository(DB)
	methodService := method.NewMethodService(methodRepo)
	methodHandler := method.NewMethodHandler(methodService)

	categoryGroup := r.Group("/payment-methods")
	{
		categoryGroup.POST("", methodHandler.Create)
		categoryGroup.GET("", methodHandler.GetAll)
		// categoryGroup.GET("/:id", methodHandler.GetSubCategoryByID)
		categoryGroup.PUT("/:id", methodHandler.Update)
		categoryGroup.DELETE("/:id", methodHandler.Delete)
	}
}
