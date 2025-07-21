package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type CreateTransaction struct {
	ProductCode string  `json:"productCode"`
	MethodCode  string  `json:"methodCode"`
	WhatsApp    string  `json:"whatsapp"`
	Username    *string `json:"username"`
	Role        string  `json:"role"`
	VoucherCode *string `json:"voucherCode,omitempty"`
	GameId      string  `json:"gameId"`
	Zone        *string `json:"zone,omitempty"`
}

type CreateTransactionResponse struct {
	OrderID string `json:"orderId"`
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (repo *TransactionRepository) Create(c context.Context, req CreateTransaction) (*CreateTransactionResponse, error) {
	duitkuService := lib.NewDuitkuService(
		config.GetEnv("DUITKU_MERCHANT_CODE", ""),
		config.GetEnv("DUITKU_API_KEY", ""),
	)

	var (
		userPrice      int
		userProfit     int
		discount       int
		fee            int
		feeType        string
		methodName     string
		total          int
		price          int
		pricePlatinum  int
		isProfitFixed  string
		priceReseller  int
		pricePurchase  int
		profit         int
		profitPlatinum int
		profitReseller int
		providerID     string
	)

	query := `
		SELECT 
			price,
			price_platinum,
			price_reseller,
			price_purchase,
			profit,
			profit_platinum,
			profit_reseller,
			provider_id,
			is_profit_fixed
		FROM services
		WHERE provider_id = $1
	`
	row := repo.db.QueryRowContext(c, query, req.ProductCode)

	err := row.Scan(&price, &pricePlatinum, &priceReseller, &pricePurchase,
		&profit, &profitPlatinum, &profitReseller, &providerID, &isProfitFixed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to query service: %w", err)
	}

	switch strings.ToUpper(req.Role) {
	case "PLATINUM":
		userPrice = pricePlatinum
		if isProfitFixed == "active" {
			userProfit = profitPlatinum
		} else {
			userProfit = pricePlatinum - pricePurchase
		}
	case "RESELLER":
		userPrice = priceReseller
		if isProfitFixed == "active" {
			userProfit = profitReseller
		} else {
			userProfit = priceReseller - pricePurchase
		}
	default:
		userPrice = price
		if isProfitFixed == "active" {
			userProfit = profit
		} else {
			userProfit = price - pricePurchase
		}
	}

	if req.VoucherCode != nil && *req.VoucherCode != "" {
		calculatedDiscount, err := repo.calculateVoucherDiscount(c, *req.VoucherCode, userPrice)
		if err != nil {
			return nil, fmt.Errorf("voucher error: %w", err)
		}

		discount = calculatedDiscount
		userPrice = userPrice - discount
		userProfit = userPrice - pricePurchase
		if userProfit < 0 {
			userProfit = 0
		}
	}

	if req.MethodCode != "SALDO" {
		calculatedFee, methodNameResult, err := repo.calculatePaymentFee(c, req.MethodCode, userPrice)
		if err != nil {
			return nil, fmt.Errorf("payment method error: %w", err)
		}

		fee = calculatedFee
		methodName = methodNameResult
	}

	total = userPrice + fee

	orderID, err := repo.insertTransaction(c, req, userPrice, discount, fee, total, userProfit)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if req.VoucherCode != nil && *req.VoucherCode != "" {
		err = repo.updateVoucherUsage(c, *req.VoucherCode)
		if err != nil {
			return nil, nil
		}
	}

	response, err := duitkuService.CreateTransaction(c, &lib.DuitkuCreateTransactionParams{
		PaymentAmount:   userPrice,
		MerchantOrderId: orderID,
		ProductDetails:  "",
		PaymentCode:     req.MethodCode,
		Cust:            stringPtr(req.MethodCode),
		CallbackUrl:     stringPtr(config.GetEnv("DUITKU_CALLBACK_URL", "")),
		ReturnUrl:       stringPtr(config.GetEnv("DUITKU_RETURN_URL", "")),
		NoWa:            req.WhatsApp,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)

	}

	fmt.Println(response, methodName, feeType)

	return &CreateTransactionResponse{
		OrderID: orderID,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}

func (repo *TransactionRepository) calculateVoucherDiscount(c context.Context, voucherCode string, userPrice int) (int, error) {
	var (
		discountType  string
		discountValue float64
		maxDiscount   sql.NullFloat64
		minPurchase   sql.NullFloat64
		usageLimit    sql.NullInt64
		usageCount    sql.NullInt64
		startDate     sql.NullTime
		expiryDate    sql.NullTime
		isActive      string
		voucherId     int
	)

	voucherQuery := `
		SELECT id, discount_type, discount_value, max_discount, min_purchase,
			   usage_limit, usage_count, start_date, expiry_date, is_active
		FROM vouchers
		WHERE code = $1
	`

	err := repo.db.QueryRowContext(c, voucherQuery, voucherCode).Scan(
		&voucherId, &discountType, &discountValue, &maxDiscount, &minPurchase,
		&usageLimit, &usageCount, &startDate, &expiryDate, &isActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("voucher not found")
		}
		return 0, fmt.Errorf("failed to query voucher: %w", err)
	}

	// Validasi voucher
	now := time.Now()
	if isActive != "active" {
		return 0, fmt.Errorf("voucher is not active")
	}

	if startDate.Valid && now.Before(startDate.Time) {
		return 0, fmt.Errorf("voucher not yet valid")
	}

	if expiryDate.Valid && now.After(expiryDate.Time) {
		return 0, fmt.Errorf("voucher has expired")
	}

	if usageLimit.Valid && usageCount.Valid && usageCount.Int64 >= usageLimit.Int64 {
		return 0, fmt.Errorf("voucher usage limit reached")
	}

	if minPurchase.Valid && float64(userPrice) < minPurchase.Float64 {
		return 0, fmt.Errorf("minimum purchase amount not met for voucher")
	}

	// Hitung diskon
	var discount int
	switch strings.ToUpper(discountType) {
	case "PERCENTAGE":
		discount = int(float64(userPrice) * (discountValue / 100))
	case "FIXED":
		discount = int(discountValue)
	default:
		return 0, fmt.Errorf("invalid discount type: %s", discountType)
	}

	if maxDiscount.Valid && float64(discount) > maxDiscount.Float64 {
		discount = int(maxDiscount.Float64)
	}

	if discount > userPrice {
		discount = userPrice
	}

	return discount, nil
}

func (repo *TransactionRepository) calculatePaymentFee(c context.Context, methodCode string, userPrice int) (int, string, error) {
	var (
		feeValue   float64
		feeType    string
		methodName string
	)

	queryFee := `
		SELECT 
			fee,
			fee_type,
			name
		FROM payment_methods
		WHERE code = $1 AND status = 'active'
	`

	err := repo.db.QueryRowContext(c, queryFee, methodCode).Scan(&feeValue, &feeType, &methodName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("payment method not found")
		}
		return 0, "", fmt.Errorf("failed to query payment method: %w", err)
	}

	var calculatedFee int
	switch strings.ToUpper(feeType) {
	case "PERCENTAGE":
		calculatedFee = utils.CalculateFeeQris(userPrice)
	case "FIXED":
		calculatedFee = int(feeValue)
	default:
		return 0, "", fmt.Errorf("invalid fee type: %s", feeType)
	}

	return calculatedFee, methodName, nil
}

func (repo *TransactionRepository) insertTransaction(c context.Context, req CreateTransaction,
	userPrice, discount, fee, total, userProfit int) (string, error) {

	prefix := "VAZZ"
	orderID := utils.GenerateUniqeID(&prefix)

	tx, err := repo.db.BeginTx(c, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var serviceName string
	serviceQuery := `SELECT service_name FROM services WHERE provider_id = $1`
	err = tx.QueryRowContext(c, serviceQuery, req.ProductCode).Scan(&serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get service name: %w", err)
	}

	insertTransactionQuery := `
		INSERT INTO transactions (
			order_id, username, purchase_price, discount, user_id, zone, 
			service_name, price, profit, profit_amount, status, is_digi, 
			success_report_sent, transaction_type
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	username := "adminaja"
	_, err = tx.ExecContext(c, insertTransactionQuery,
		orderID,
		username,
		userPrice,
		discount,
		req.GameId,
		req.Zone,
		serviceName,
		userPrice,
		userProfit,
		userProfit,
		"PENDING",
		"active",
		"active",
		"TOPUP",
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert transaction: %w", err)
	}

	insertPaymentQuery := `
		INSERT INTO payments (
			order_id, price, total_amount, buyer_number, fee, 
			fee_amount, status, method
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err = tx.ExecContext(c, insertPaymentQuery,
		orderID,
		fmt.Sprintf("%d", userPrice),
		total,
		req.WhatsApp,
		fee,
		fee,
		"PENDING",
		req.MethodCode,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert payment: %w", err)
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orderID, nil
}

func (repo *TransactionRepository) updateVoucherUsage(c context.Context, voucherCode string) error {
	updateQuery := `
		UPDATE vouchers 
		SET usage_count = COALESCE(usage_count, 0) + 1,
			updated_at = NOW()
		WHERE code = $1
	`

	_, err := repo.db.ExecContext(c, updateQuery, voucherCode)
	if err != nil {
		return fmt.Errorf("failed to update voucher usage: %w", err)
	}

	return nil
}
