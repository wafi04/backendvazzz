package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/service/analytics"
)

func SetupAnalyticsRoutes(router *gin.RouterGroup, db *sql.DB) {
	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo)
	analyticsHandler := analytics.NewAnalyticsHandler(analyticsService)

	analyticsGroup := router.Group("/analytics")
	// analyticsGroup.Use(middleware.AdminMiddleware())
	{
		analyticsGroup.GET("/range", analyticsHandler.GetAnalyticsByDateRange)
		analyticsGroup.GET("/date", analyticsHandler.GetAnalyticsByDate)
		analyticsGroup.GET("/today", analyticsHandler.GetTodayAnalytics)
		analyticsGroup.GET("/monthly", analyticsHandler.GetMonthlyAnalytics)
		analyticsGroup.GET("/summary", analyticsHandler.GetAnalyticsSummary)
		analyticsGroup.GET("/status", analyticsHandler.GetAnalyticsByStatus)
	}
}
