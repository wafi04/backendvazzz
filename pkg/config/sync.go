package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/service/product"
)

type SyncManager struct {
	Ticker      *time.Ticker
	Digi        *lib.DigiflazzService
	ProductRepo *product.ProductRepository
	IsRunning   bool
	IsSyncing   bool // Track if sync is currently in progress
	Mutex       sync.RWMutex
	SyncMutex   sync.Mutex // Separate mutex for sync operations
	StopChan    chan struct{}
	Db          *sql.DB
}

// Global sync manager instance with proper initialization
var (
	GlobalSyncManager *SyncManager
	SyncManagerMutex  sync.Mutex
)

// NewSyncManager creates a new sync manager
func NewSyncManager(db *sql.DB) *SyncManager {
	digi := lib.NewDigiflazzService(lib.DigiConfig{
		DigiKey:      GetEnv("DIGI__API_KEY", ""),
		DigiUsername: GetEnv("DIGI_USERNAME", ""),
	})

	return &SyncManager{
		Digi:        digi,
		ProductRepo: product.NewProductRepository(db),
		Db:          db,
		StopChan:    make(chan struct{}),
		IsRunning:   false,
		IsSyncing:   false,
	}
}

// GetGlobalSyncManager returns the global sync manager instance
func GetGlobalSyncManager() *SyncManager {
	SyncManagerMutex.Lock()
	defer SyncManagerMutex.Unlock()
	return GlobalSyncManager
}

// SetGlobalSyncManager sets the global sync manager instance
func SetGlobalSyncManager(sm *SyncManager) {
	SyncManagerMutex.Lock()
	defer SyncManagerMutex.Unlock()
	GlobalSyncManager = sm
}

// Start begins the periodic sync process
func (sm *SyncManager) Start() {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	if sm.IsRunning {
		return
	}

	sm.Ticker = time.NewTicker(10 * time.Minute)
	sm.IsRunning = true

	go sm.syncLoop()
	log.Println("Product sync started - running every 10 minutes")
}

// Stop stops the periodic sync process
func (sm *SyncManager) Stop() {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	if !sm.IsRunning {
		return
	}

	sm.Ticker.Stop()
	close(sm.StopChan)
	sm.IsRunning = false
	log.Println("Product sync stopped")
}

// GetIsRunning checks if sync is currently running
func (sm *SyncManager) GetIsRunning() bool {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()
	return sm.IsRunning
}

// GetIsSyncing checks if sync is currently in progress
func (sm *SyncManager) GetIsSyncing() bool {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()
	return sm.IsSyncing
}

// SetSyncingStatus sets the syncing status (for external control if needed)
func (sm *SyncManager) SetSyncingStatus(status bool) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	sm.IsSyncing = status
}

// syncLoop runs the periodic sync
func (sm *SyncManager) syncLoop() {
	defer func() {
		sm.Mutex.Lock()
		sm.IsRunning = false
		sm.Mutex.Unlock()
	}()

	// Run initial sync
	sm.PerformSync()

	for {
		select {
		case <-sm.Ticker.C:
			sm.PerformSync()
		case <-sm.StopChan:
			return
		}
	}
}

// PerformSync executes the actual sync process with memory optimization
func (sm *SyncManager) PerformSync() {
	// Use separate mutex to prevent concurrent syncs
	sm.SyncMutex.Lock()
	defer sm.SyncMutex.Unlock()

	// Set syncing state
	sm.Mutex.Lock()
	if sm.IsSyncing {
		sm.Mutex.Unlock()
		log.Println("Sync already in progress, skipping...")
		return
	}
	sm.IsSyncing = true
	sm.Mutex.Unlock()

	// Always reset syncing state when done
	defer func() {
		sm.Mutex.Lock()
		sm.IsSyncing = false
		sm.Mutex.Unlock()
	}()

	log.Println("Starting product sync...")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Fetch products with context for timeout
	products, err := sm.Digi.CheckPrice()
	if err != nil {
		log.Printf("Failed to fetch price list: %v", err)
		return
	}

	if len(products) == 0 {
		log.Println("No products found from API")
		return
	}

	// Process products in batches to save memory
	batchSize := 100 // Process 100 products at a time
	totalProducts := len(products)
	successCount := 0
	errorCount := 0

	for i := 0; i < totalProducts; i += batchSize {
		// Check if we should stop (context cancelled or stop signal)
		select {
		case <-ctx.Done():
			log.Printf("Sync cancelled due to timeout")
			return
		default:
		}

		// Check if sync manager is still running
		if !sm.GetIsRunning() && sm != GlobalSyncManager {
			log.Println("Sync manager stopped, aborting sync...")
			return
		}

		end := i + batchSize
		if end > totalProducts {
			end = totalProducts
		}

		batch := products[i:end]
		batchSuccess, batchError := sm.ProcessBatch(ctx, batch)
		successCount += batchSuccess
		errorCount += batchError

		// Force garbage collection after each batch to free memory
		batch = nil

		// Small delay between batches to prevent overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	log.Printf("Sync completed in %v - Success: %d, Errors: %d, Total: %d",
		duration, successCount, errorCount, totalProducts)

	// Clear the products slice to free memory
	products = nil
}

// ProcessBatch processes a batch of products
func (sm *SyncManager) ProcessBatch(ctx context.Context, batch []*lib.ProductData) (int, int) {
	successCount := 0
	errorCount := 0

	for _, product := range batch {
		if product == nil {
			errorCount++
			continue
		}

		// Check if product exists and update/create accordingly
		exists, err := sm.ProductExists(ctx, product.BuyerSkuCode)
		if err != nil {
			log.Printf("Error checking product existence for %s: %v", product.ProductName, err)
			errorCount++
			continue
		}

		if exists {
			// Update existing product
			err = sm.ProductRepo.UpdatePrice(ctx, product)
		} else {
			// Create new product
			err = sm.ProductRepo.Create(ctx, *product)
		}

		if err != nil {
			log.Printf("Failed to process product %s: %v", product.ProductName, err)
			errorCount++
		} else {
			successCount++
		}
	}

	return successCount, errorCount
}

// ProductExists checks if a product already exists in database
func (sm *SyncManager) ProductExists(ctx context.Context, providerID string) (bool, error) {
	query := `SELECT 1 FROM services WHERE provider_id = $1 LIMIT 1`
	var exists int
	err := sm.Db.QueryRowContext(ctx, query, providerID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ManualSync allows external packages to trigger a manual sync
func (sm *SyncManager) ManualSync() error {
	if sm.GetIsSyncing() {
		return fmt.Errorf("sync is already in progress")
	}

	go sm.PerformSync()
	return nil
}

// GetSyncStatus returns the current sync status information
func (sm *SyncManager) GetSyncStatus() map[string]interface{} {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()

	return map[string]interface{}{
		"is_running":    sm.IsRunning,
		"is_syncing":    sm.IsSyncing,
		"ticker_active": sm.Ticker != nil,
	}
}
