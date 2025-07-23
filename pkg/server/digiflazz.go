package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/config"
)

func getSyncManager(db *sql.DB) *config.SyncManager {
	config.SyncManagerMutex.Lock()
	defer config.SyncManagerMutex.Unlock()

	if config.GlobalSyncManager == nil {
		config.GlobalSyncManager = config.NewSyncManager(db)
	}
	return config.GlobalSyncManager
}

func GetProductFromDigiflazz(r *gin.RouterGroup, db *sql.DB) {
	syncManager := getSyncManager(db)

	digRoutes := r.Group("/sync")

	digRoutes.GET("/", func(c *gin.Context) {
		if syncManager.IsSyncing {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Sync already in progress",
				"message": "Please wait for current sync to complete",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
		defer cancel()

		syncManager.SyncMutex.Lock()
		syncManager.Mutex.Lock()
		if syncManager.IsSyncing {
			syncManager.Mutex.Unlock()
			syncManager.SyncMutex.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Sync already in progress",
				"message": "Please wait for current sync to complete",
			})
			return
		}
		syncManager.IsSyncing = true
		syncManager.Mutex.Unlock()

		defer func() {
			syncManager.Mutex.Lock()
			syncManager.IsSyncing = false
			syncManager.Mutex.Unlock()
			syncManager.SyncMutex.Unlock()
		}()

		products, err := syncManager.Digi.CheckPrice()
		if err != nil {
			log.Printf("Failed to fetch price list: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch price list",
				"details": err.Error(),
			})
			return
		}

		if len(products) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "No products found",
				"count":   0,
			})
			return
		}

		batchSize := 50
		successCount := 0
		errorCount := 0
		var errors []string

		for i := 0; i < len(products); i += batchSize {
			select {
			case <-ctx.Done():
				c.JSON(http.StatusRequestTimeout, gin.H{
					"error":   "Sync timeout",
					"message": "Sync cancelled due to timeout",
				})
				return
			default:
			}

			end := i + batchSize
			if end > len(products) {
				end = len(products)
			}

			batch := products[i:end]
			for _, product := range batch {
				if product == nil {
					errorCount++
					continue
				}

				exists, err := syncManager.ProductExists(ctx, product.BuyerSkuCode)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Error checking %s: %v", product.ProductName, err))
					errorCount++
					continue
				}

				if exists {
					err = syncManager.ProductRepo.UpdatePrice(ctx, product)
				} else {
					err = syncManager.ProductRepo.Create(ctx, *product)
				}

				if err != nil {
					errors = append(errors, fmt.Sprintf("Failed to process %s: %v", product.ProductName, err))
					errorCount++
				} else {
					successCount++
				}
			}
		}

		response := gin.H{
			"status":        "success",
			"message":       "Product sync completed",
			"total_fetched": len(products),
			"success_count": successCount,
			"error_count":   errorCount,
		}

		if len(errors) > 0 {
			response["errors"] = errors
		}

		c.JSON(http.StatusOK, response)
	})

	digRoutes.POST("/start", func(c *gin.Context) {
		if syncManager.IsRunning {
			c.JSON(http.StatusOK, gin.H{
				"status":  "already_running",
				"message": "Automatic sync is already running",
			})
			return
		}

		syncManager.Start()
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Automatic sync started - will run every 10 minutes",
		})
	})

	digRoutes.POST("/stop", func(c *gin.Context) {
		if !syncManager.IsRunning {
			c.JSON(http.StatusOK, gin.H{
				"status":  "not_running",
				"message": "Automatic sync is not running",
			})
			return
		}

		syncManager.Stop()
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Automatic sync stopped",
		})
	})

	digRoutes.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"is_running": syncManager.IsRunning,
			"is_syncing": syncManager.IsSyncing,
			"interval":   "10 minutes",
		})
	})
}
