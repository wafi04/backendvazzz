package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type CreateTransaction struct {
	ProductCode string  `json:"productCode"`
	MethodCode  string  `json:"methodCode"`
	WhatsApp    string  `json:"whatsapp"`
	Username    *string `json:"username,omitempty"`
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

	prefix := "VAZZ"
	orderID := utils.GenerateUniqeID(&prefix)

	tx, err := repo.db.BeginTx(c, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && rErr != sql.ErrTxDone {
			fmt.Printf("Error during transaction rollback: %v\n", rErr)
		}
	}()

	var (
		userPrice      int
		userProfit     int
		discount       int
		fee            int
		methodName     string
		feeType        string
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
	row := tx.QueryRowContext(c, query, req.ProductCode)

	err = row.Scan(&price, &pricePlatinum, &priceReseller, &pricePurchase,
		&profit, &profitPlatinum, &profitReseller, &providerID, &isProfitFixed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found for product code '%s'", req.ProductCode)
		}
		return nil, fmt.Errorf("failed to query service details: %w", err)
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
		calculatedDiscount, err := repo.calculateVoucherDiscount(c, tx, *req.VoucherCode, userPrice)
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
		calculatedFee, methodNameResult, err := repo.calculatePaymentFee(c, tx, req.MethodCode, userPrice)
		if err != nil {
			return nil, fmt.Errorf("payment method error: %w", err)
		}

		fee = calculatedFee
		methodName = methodNameResult
	} else {
		if req.Username == nil || *req.Username == "" {
			return nil, fmt.Errorf("username is required for SALDO payment")
		}

		total = userPrice + fee
		_, err = repo.PaymentUsingSaldo(c, CreatePaymentUsingSaldo{
			Username: *req.Username,
			OrderID:  orderID,
			Total:    total,
			WhatsApp: req.WhatsApp,
			FeeType:  feeType,
			Fee:      fee,
			Price:    userPrice,
			Tx:       tx,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to process saldo payment: %w", err)
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit SALDO transaction: %w", err)
		}

		return &CreateTransactionResponse{
			OrderID: orderID,
		}, nil
	}

	total = userPrice + fee

	err = repo.insertTransaction(c, tx, req, orderID, userPrice, discount, fee, total, userProfit, methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction records: %w", err)
	}

	if req.VoucherCode != nil && *req.VoucherCode != "" {
		err = repo.updateVoucherUsage(c, *req.VoucherCode, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to update voucher usage: %w", err)
		}
	}

	_, err = duitkuService.CreateTransaction(c, &lib.DuitkuCreateTransactionParams{
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
		return nil, fmt.Errorf("failed to create external payment with Duitku: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CreateTransactionResponse{
		OrderID: orderID,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}

func (repo *TransactionRepository) insertTransaction(c context.Context, tx *sql.Tx, req CreateTransaction,
	orderID string, userPrice, discount, fee, total, userProfit int, methodName string) error {

	var serviceName string
	var purchasePrice int
	serviceQuery := `SELECT service_name,price_purchase FROM services WHERE provider_id = $1`
	err := tx.QueryRowContext(c, serviceQuery, req.ProductCode).Scan(&serviceName, &purchasePrice)
	if err != nil {
		return fmt.Errorf("failed to get service name: %w", err)
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

	transactionUsername := "adminaja"
	if req.Username != nil && *req.Username != "" {
		transactionUsername = *req.Username
	}

	_, err = tx.ExecContext(c, insertTransactionQuery,
		orderID,
		transactionUsername,
		purchasePrice,
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
		return fmt.Errorf("failed to insert transaction record: %w", err)
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
		methodName,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment record: %w", err)
	}

	return nil
}

func (repo *TransactionRepository) updateVoucherUsage(c context.Context, voucherCode string, tx *sql.Tx) error {
	updateQuery := `
        UPDATE vouchers
        SET usage_count = COALESCE(usage_count, 0) + 1,
            updated_at = NOW()
        WHERE code = $1
    `

	_, err := tx.ExecContext(c, updateQuery, voucherCode)
	if err != nil {
		return fmt.Errorf("failed to update voucher usage: %w", err)
	}

	return nil
}
