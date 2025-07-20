package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	middleware "github.com/wafi04/backendvazzz/pkg/midlleware"
	"github.com/wafi04/backendvazzz/service/users"
)

func SetupRoutesUser(api *gin.RouterGroup, db *sql.DB) {
	userRepo := users.NewUserRepository(db)
	userService := users.NewUserService(userRepo)
	authHandler := users.NewAuthHandler(userService)
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("/auth")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/logout", authHandler.Logout)
		}
	}
}
