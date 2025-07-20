package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/service/news"
)

func SetupRoutesNews(r *gin.RouterGroup, DB *sql.DB) {
	newsRepo := news.NewNewsRepository(DB)
	newsService := news.NewNewsService(newsRepo)
	newsHandler := news.NewNewsHandler(newsService)

	categoryGroup := r.Group("/news")
	{
		categoryGroup.POST("", newsHandler.Create)
		categoryGroup.GET("", newsHandler.GetAll)
		// categoryGroup.GET("/:id", newsHandler.GetSubCategoryByID)
		categoryGroup.PUT("/:id", newsHandler.Update)
		categoryGroup.DELETE("/:id", newsHandler.Delete)
	}
}
