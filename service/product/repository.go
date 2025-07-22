package product

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/model"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

// ProfitConfig holds the default profit configuration
type ProfitConfig struct {
	Profit         int
	ProfitReseller int
	ProfitPlatinum int
	IsProfitFixed  string
}

// getDefaultProfitConfig returns default profit configuration based on category
func getDefaultProfitConfig(category string) ProfitConfig {
	isFixed := category == "Voucher" || category == "PLN"

	isProfitFixedStr := "inactive"
	if isFixed {
		isProfitFixedStr = "active"
	}

	return ProfitConfig{
		Profit:         4,
		ProfitReseller: 3,
		ProfitPlatinum: 2,
		IsProfitFixed:  isProfitFixedStr,
	}
}

// extractProviderFromSKU extracts provider code from SKU using regex
func extractProviderFromSKU(buyerSkuCode string) string {
	re := regexp.MustCompile(`^([A-Z]+)`)
	match := re.FindStringSubmatch(buyerSkuCode)
	if len(match) > 1 {
		return match[1]
	}
	return buyerSkuCode
}

func (repo *ProductRepository) getCategoryAndSubCategory(ctx context.Context, categoryName, brand string) (int, int, error) {
	var categoryID sql.NullInt64

	queryCat := `SELECT id FROM categories WHERE LOWER(brand) = LOWER($1)`
	err := repo.DB.QueryRowContext(ctx, queryCat, brand).Scan(&categoryID)
	if err != nil {
		queryCatName := `SELECT id FROM categories WHERE LOWER(name) = LOWER($1)`
		err = repo.DB.QueryRowContext(ctx, queryCatName, categoryName).Scan(&categoryID)
		if err != nil {
			return 0, 0, err
		}
	}
	return int(categoryID.Int64), 1, nil
}

func calculatePrices(basePrice int, config ProfitConfig) (int, int, int, int) {
	priceFromDigi := basePrice

	var priceReseller, pricePlatinum, price int

	if config.IsProfitFixed == "active" {
		// Fixed profit
		priceReseller = priceFromDigi + config.ProfitReseller
		pricePlatinum = priceFromDigi + config.ProfitPlatinum
		price = priceFromDigi + config.Profit
	} else {
		priceReseller = priceFromDigi + ((priceFromDigi * config.ProfitReseller) / 100)
		pricePlatinum = priceFromDigi + ((priceFromDigi * config.ProfitPlatinum) / 100)
		price = priceFromDigi + ((priceFromDigi * config.Profit) / 100)
	}

	return price, priceReseller, pricePlatinum, priceFromDigi
}

func (repo *ProductRepository) Create(ctx context.Context, req lib.ProductData) error {

	categoryID, subCategoryID, err := repo.getCategoryAndSubCategory(ctx, req.Category, req.Brand)
	if err != nil {
		return err
	}

	profitConfig := getDefaultProfitConfig(req.Category)

	price, priceReseller, pricePlatinum, priceFromDigi := calculatePrices(req.Price, profitConfig)

	query := `
		INSERT INTO services (
			service_name,
			category_id,
			sub_category_id,
			price,
			price_purchase,
			price_reseller,
			price_platinum,
			price_suggest,
			profit,
			profit_platinum,
			profit_reseller,
			profit_suggest,
			is_suggest,
			status,
			provider_id,
			provider,
			note,
			is_profit_fixed,
			product_logo,
			is_flash_sale,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW(), NOW())
	`

	// Determine status based on seller product status
	status := "inactive"
	if req.SellerProductStatus {
		status = "active"
	}

	// Use description as note
	note := req.Desc
	if note == "" {
		note = ""
	}

	// Handle NULL sub_category_id
	var subCategoryParam interface{}
	if subCategoryID == 0 {
		subCategoryParam = nil
	} else {
		subCategoryParam = subCategoryID
	}

	_, err = repo.DB.ExecContext(ctx, query,
		req.ProductName,             // $1 - service_name
		categoryID,                  // $2 - category_id
		subCategoryParam,            // $3 - sub_category_id
		price,                       // $4 - price
		priceFromDigi,               // $5 - price_from_digi
		priceReseller,               // $6 - price_reseller
		pricePlatinum,               // $7 - price_platinum
		0,                           // $8 - price_suggest
		profitConfig.Profit,         // $9 - profit
		profitConfig.ProfitPlatinum, // $10 - profit_platinum
		profitConfig.ProfitReseller, // $11 - profit_reseller
		0,                           // $12 - profit_suggest
		"inactive",                  // $13 - is_suggest
		status,                      // $14 - status
		req.BuyerSkuCode,            // $15 - provider_id
		"digiflazz",                 // $16 - provider
		note,                        // $17 - note
		profitConfig.IsProfitFixed,  // $18 - is_profit_fixed
		nil,                         // $19 - product_logo
		"inactive",                  // $20 - is_flash_sale
	)

	return err
}

func (repo *ProductRepository) UpdatePrice(ctx context.Context, req *lib.ProductData) error {
	var existingProfit, existingProfitReseller, existingProfitPlatinum int
	var isProfitFixed string

	queryExisting := `
		SELECT profit, profit_reseller, profit_platinum, is_profit_fixed 
		FROM services 
		WHERE provider_id = $1
	`

	err := repo.DB.QueryRowContext(ctx, queryExisting, req.BuyerSkuCode).Scan(
		&existingProfit, &existingProfitReseller, &existingProfitPlatinum, &isProfitFixed,
	)
	if err != nil {
		// If product doesn't exist, fallback to default config
		profitConfig := getDefaultProfitConfig(req.Category)
		existingProfit = profitConfig.Profit
		existingProfitReseller = profitConfig.ProfitReseller
		existingProfitPlatinum = profitConfig.ProfitPlatinum
		isProfitFixed = profitConfig.IsProfitFixed
	}

	// Use existing profit configuration
	config := ProfitConfig{
		Profit:         existingProfit,
		ProfitReseller: existingProfitReseller,
		ProfitPlatinum: existingProfitPlatinum,
		IsProfitFixed:  isProfitFixed,
	}

	// Calculate new prices with existing profit configuration
	price, priceReseller, pricePlatinum, priceFromDigi := calculatePrices(req.Price, config)

	status := "inactive"
	if req.SellerProductStatus {
		status = "active"
	}

	// Update query - only update price-related fields and status
	updateQuery := `
		UPDATE services 
		SET 
			price = $1,
			price_purchase = $2,
			price_reseller = $3,
			price_platinum = $4,
			status = $5,
			updated_at = $6
		WHERE provider_id = $7
	`

	_, err = repo.DB.ExecContext(ctx, updateQuery,
		price,            // $1
		priceFromDigi,    // $2
		priceReseller,    // $3
		pricePlatinum,    // $4
		status,           // $5
		time.Now(),       // $6
		req.BuyerSkuCode, // $7
	)

	if err != nil {
		return err
	}

	println("DEBUG: Updated prices for product:", req.ProductName, "SKU:", req.BuyerSkuCode)
	return nil
}

func (repo *ProductRepository) GetExistingProductCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM services WHERE provider = 'digiflazz'`
	err := repo.DB.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

type ProductWithUserPrice struct {
	ServiceName   string    `json:"serviceName"`
	CategoryID    int       `json:"categoryId"`
	SubCategoryID int       `json:"subCategoryId"`
	UserPrice     int       `json:"userPrice"`
	UserProfit    int       `json:"userProfit"`
	IsSuggest     string    `json:"isSuggest"`
	SuggestPrice  *int      `json:"suggestPrice,omitempty"`
	ProvideId     string    `json:"providerId"`
	ProductLogo   *string   `json:"productLogo,omitempty"`
	IsFlashSale   string    `json:"isFlashSale"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (repo *ProductRepository) GetAll(categoryId, subCategoryID int, role string) ([]ProductWithUserPrice, error) {
	baseQuery := `
	SELECT 
		service_name,
		category_id,
		sub_category_id,
		price,
		price_purchase,
		price_reseller,
		price_platinum,
		price_suggest,
		profit,
		profit_platinum,
		profit_reseller,
		profit_suggest,
		is_suggest,
		provider_id,
		product_logo,
		is_flash_sale,
		created_at,
		updated_at
	FROM services
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if categoryId > 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, categoryId)
		argIndex++
	}

	if subCategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("sub_category_id = $%d", argIndex))
		args = append(args, subCategoryID)
		argIndex++
	}

	// Hanya ambil yang status = active
	conditions = append(conditions, "status = 'active'")

	// Final query
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY price ASC"

	// Query ke DB
	rows, err := repo.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var products []ProductWithUserPrice

	for rows.Next() {
		var (
			serviceName                            string
			categoryID, subCategoryID              int
			price, pricePurchase                   int
			priceReseller, pricePlatinum           int
			priceSuggest                           int
			profit, profitPlatinum, profitReseller int
			profitSuggest                          int
			isSuggest                              string
			providerId                             string
			productLogo                            *string
			isFlashSale                            string
			createdAt, updatedAt                   time.Time
		)

		err := rows.Scan(
			&serviceName,
			&categoryID,
			&subCategoryID,
			&price,
			&pricePurchase,
			&priceReseller,
			&pricePlatinum,
			&priceSuggest,
			&profit,
			&profitPlatinum,
			&profitReseller,
			&profitSuggest,
			&isSuggest,
			&providerId,
			&productLogo,
			&isFlashSale,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Hitung user price berdasarkan role
		product := model.Services{
			Price:          price,
			PricePurchase:  pricePurchase,
			PriceReseller:  priceReseller,
			PricePlatinum:  pricePlatinum,
			Profit:         profit,
			ProfitPlatinum: profitPlatinum,
			ProfitReseller: profitReseller,
		}

		userPrice, userProfit := repo.calculateUserPriceAndProfit(product, role)

		var suggestPrice *int
		if isSuggest == "active" {
			suggestPrice = &priceSuggest
		}

		products = append(products, ProductWithUserPrice{
			ServiceName:   serviceName,
			CategoryID:    categoryID,
			SubCategoryID: subCategoryID,
			UserPrice:     userPrice,
			UserProfit:    userProfit,
			IsSuggest:     isSuggest,
			SuggestPrice:  suggestPrice,
			ProvideId:     providerId,
			ProductLogo:   productLogo,
			IsFlashSale:   isFlashSale,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, nil
}

func (repo *ProductRepository) calculateUserPriceAndProfit(product model.Services, role string) (int, int) {
	switch strings.ToUpper(role) {
	case "ADMIN":
		return product.PricePurchase, product.Profit
	case "PLATINUM":
		return product.PricePlatinum, product.ProfitPlatinum
	case "RESELLER":
		return product.PriceReseller, product.ProfitReseller
	case "MEMBER":
		return product.Price, 0
	default:
		return product.Price, 0
	}
}
