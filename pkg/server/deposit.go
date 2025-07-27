package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	middleware "github.com/wafi04/backendvazzz/pkg/midlleware"
	"github.com/wafi04/backendvazzz/service/deposit"
)

func SetupDepositTransaction(r *gin.RouterGroup, DB *sql.DB) {

	depositRepo := deposit.NewDepositRepository(DB)
	depositService := deposit.NewDepositService(depositRepo)
	depositHandler := deposit.NewDepositHandler(depositService)

	routes := r.Group("/deposit")
	routes.Use(middleware.AuthMiddleware())
	{
		routes.POST("", depositHandler.Create)
		routes.GET("/by/username", depositHandler.GetAllByUsername)
		routes.GET("/:id", depositHandler.GetByDepositID)
		routes.DELETE("/:id", depositHandler.Delete)
		routes.GET("", depositHandler.GetAll)
	}

}
