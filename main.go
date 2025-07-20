package main

import (
	"database/sql"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/server"
)

func main() {
	config.LoadEnv()
	// Load environment variables
	dbURL := config.GetEnv("DATABASE_URL", "")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Setup Gin router
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "https://your-frontend.com"}
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	api := r.Group("/api")
	log.Println("🔧 Setting up all routes...")

	server.SetupRoutes(api, db)
	log.Println("✅ SetupRoutes completed")

	server.SetupRoutesSubCategories(api, db)
	log.Println("✅ SetupRoutesSubCategories completed")

	server.SetupRoutesMethod(api, db)
	log.Println("✅ SetupRoutesMethod completed")

	server.GetProductFromDigiflazz(api, db)
	log.Println("✅ GetProductFromDigiflazz completed")

	server.SetupRoutesUser(api, db)
	log.Println("✅ SetupRoutesUser completed")

	server.SetupRoutesNews(api, db)
	log.Println("✅ SetupRoutesNews completed")

	server.SetupRoutesProducts(api, db)
	log.Println("✅ SetupRoutesProducts completed")

	server.SetUpTransactionRoutes(api, db)
	log.Println("✅ SetUpTransactionRoutes completed")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Server is running",
		})
	})

	// Debug endpoint untuk list semua routes
	r.GET("/routes", func(c *gin.Context) {
		routes := []gin.H{}
		for _, route := range r.Routes() {
			routes = append(routes, gin.H{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(200, gin.H{
			"message": "Available routes",
			"routes":  routes,
		})
	})

	// Start server
	port := "8080"
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📍 Available endpoints:")
	log.Printf("   • Health: http://localhost:%s/health", port)
	log.Printf("   • Routes: http://localhost:%s/routes", port)
	log.Printf("   • Transaction Test: http://localhost:%s/api/transaction/test", port)
	log.Printf("   • Transaction Create: POST http://localhost:%s/api/transaction", port)

	log.Fatal(r.Run(":" + port))
}
