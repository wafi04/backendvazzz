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

type TransactionRepository struct {
	db            *sql.DB
	duitkuService *lib.DuitkuService
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	duitkuService := lib.NewDuitkuService()

	return &TransactionRepository{
		db:            db,
		duitkuService: duitkuService,
	}
}

func (repo *TransactionRepository) Create(ctx context.Context, req CreateTransaction) (*CreateTransactionResponse, error) {

	orderID := utils.GenerateUniqeID(stringPtr("VAZZ"))
	var role *string
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer repo.rollbackOnError(tx)

	if req.Username == "" {
		queryUser := `
			SELECT role
			FROM users 
			WHERE username = $1
		`
		var userRole string
		err = tx.QueryRowContext(ctx, queryUser, req.Username).Scan(&userRole)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to query user role: %w", err)
		}
		if err == nil {
			role = &userRole
		}
	}

	if role == nil {
		defaultRole := "GUEST"
		role = &defaultRole
	}
	service, err := repo.getServiceByProviderID(ctx, tx, req.ProductCode)
	if err != nil {
		return nil, err
	}

	pricing := repo.calculatePricing(service, role)

	discount := 0
	if req.VoucherCode != nil && *req.VoucherCode != "" {
		discount, err = repo.calculateVoucherDiscount(ctx, tx, *req.VoucherCode, pricing.UserPrice)
		if err != nil {
			return nil, fmt.Errorf("voucher error: %w", err)
		}

		pricing.UserPrice -= discount
		if pricing.UserProfit < 0 {
			pricing.UserProfit = 0
		}
	}

	// Handle different payment methods
	var response *CreateTransactionResponse

	if req.MethodCode == "SALDO" {
		response, err = repo.processSaldoPayment(ctx, tx, req, orderID, pricing, discount, service, req.GameId, req.Zone)
	} else {
		response, err = repo.processExternalPayment(ctx, tx, req, orderID, pricing, discount, service)
	}

	if err != nil {
		return nil, err
	}

	// Update voucher usage if applicable
	if req.VoucherCode != nil && *req.VoucherCode != "" {
		if err = repo.updateVoucherUsage(ctx, tx, *req.VoucherCode); err != nil {
			return nil, fmt.Errorf("failed to update voucher usage: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return response, nil
}

func (repo *TransactionRepository) getServiceByProviderID(ctx context.Context, tx *sql.Tx, providerID string) (*Service, error) {
	query := `
        SELECT
            price, price_platinum, price_reseller, price_purchase,
            profit, profit_platinum, profit_reseller, provider_id,
            is_profit_fixed, service_name
        FROM services
        WHERE provider_id = $1
    `

	service := &Service{}
	err := tx.QueryRowContext(ctx, query, providerID).Scan(
		&service.Price, &service.PricePlatinum, &service.PriceReseller, &service.PricePurchase,
		&service.Profit, &service.ProfitPlatinum, &service.ProfitReseller, &service.ProviderID,
		&service.IsProfitFixed, &service.ServiceName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: product code '%s'", ErrServiceNotFound, providerID)
		}
		return nil, fmt.Errorf("failed to query service details: %w", err)
	}

	return service, nil
}

func (repo *TransactionRepository) calculatePricing(service *Service, role *string) PricingResult {
	var userPrice, userProfit, userProfitAmount int

	switch strings.ToUpper(*role) {
	case "PLATINUM":
		userPrice = service.PricePlatinum
		userProfit = service.ProfitPlatinum
		userProfitAmount = service.PricePlatinum - service.PricePurchase
	case "RESELLER":
		userPrice = service.PriceReseller
		userProfit = service.ProfitReseller
		userProfitAmount = service.ProfitReseller - service.PricePurchase
	default:
		userPrice = service.Price
		userProfit = service.ProfitReseller
		userProfitAmount = service.Price - service.PricePurchase
	}

	return PricingResult{
		UserPrice:        userPrice,
		UserProfit:       userProfit,
		UserProfitAmount: userProfitAmount,
	}
}

func (repo *TransactionRepository) processSaldoPayment(ctx context.Context, tx *sql.Tx, req CreateTransaction,
	orderID string, pricing PricingResult, discount int, service *Service, userId string, zone *string) (*CreateTransactionResponse, error) {

	var NoTujuan string

	if zone != nil && *zone != "" {
		NoTujuan = fmt.Sprintf("%s%s", userId, *zone)
	} else {
		NoTujuan = userId
	}
	total := pricing.UserPrice

	// Insert transaction record
	if err := repo.insertTransaction(ctx, tx, req, orderID, pricing.UserPrice, discount, 0, total, pricing.UserProfit, pricing.UserProfitAmount, service.ServiceName); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Process saldo payment
	_, err := repo.PaymentUsingSaldo(ctx, CreatePaymentUsingSaldo{
		Username: req.Username,
		OrderID:  orderID,
		Total:    total,
		WhatsApp: req.WhatsApp,
		NoTujuan: NoTujuan,
		Price:    pricing.UserPrice,
		Tx:       tx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process saldo payment: %w", err)
	}

	return &CreateTransactionResponse{
		OrderID: orderID,
		Total:   total,
		Fee:     0,
	}, nil
}

func (repo *TransactionRepository) processExternalPayment(ctx context.Context, tx *sql.Tx, req CreateTransaction,
	orderID string, pricing PricingResult, discount int, service *Service) (*CreateTransactionResponse, error) {

	// Get payment method details
	fee, methodNameResult, err := repo.calculatePaymentFee(ctx, tx, req.MethodCode, pricing.UserPrice)
	if err != nil {
		return nil, fmt.Errorf("payment method error: %w", err)
	}

	total := pricing.UserPrice + fee

	// Insert transaction record
	if err := repo.insertTransaction(ctx, tx, req, orderID, pricing.UserPrice, discount, fee, total, pricing.UserProfit, pricing.UserProfitAmount, service.ServiceName); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Insert payment record
	if err := repo.insertPaymentRecord(ctx, tx, orderID, pricing.UserPrice, total, fee, req.WhatsApp, methodNameResult, req.MethodCode); err != nil {
		return nil, fmt.Errorf("failed to insert payment record: %w", err)
	}

	return &CreateTransactionResponse{
		OrderID: orderID,
		Total:   total,
		Fee:     fee,
	}, nil
}

func (repo *TransactionRepository) insertTransaction(ctx context.Context, tx *sql.Tx, req CreateTransaction,
	orderID string, userPrice, discount, fee, total, userProfit, profitAmount int, serviceName string) error {

	insertTransactionQuery := `
        INSERT INTO transactions (
            order_id, username,provider_order_id, purchase_price, discount, user_id, zone,
            service_name, price, profit, profit_amount, status, is_digi,
            success_report_sent, transaction_type, created_at,message
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,$15, NOW(),'Transaction Pending'
        )
    `

	_, err := tx.ExecContext(ctx, insertTransactionQuery,
		orderID,
		req.Username,
		req.ProductCode,
		userPrice,
		discount,
		req.GameId,
		req.Zone,
		serviceName,
		userPrice,
		userProfit,
		profitAmount,
		"PENDING",
		"active",
		"active",
		"TOPUP",
	)
	if err != nil {
		return fmt.Errorf("failed to insert transaction record: %w", err)
	}

	return nil
}

func (repo *TransactionRepository) insertPaymentRecord(ctx context.Context, tx *sql.Tx, orderID string, price, total, fee int, whatsApp, methodName, methodCode string) error {

	duitku, err := repo.duitkuService.CreateTransaction(ctx, &lib.DuitkuCreateTransactionParams{
		PaymentAmount:   price,
		MerchantOrderId: orderID,
		ProductDetails:  "",
		PaymentCode:     methodCode,
		Cust:            stringPtr(methodCode),
		CallbackUrl:     stringPtr(config.GetEnv("DUITKU_CALLBACK_URL", "")),
		ReturnUrl:       stringPtr(config.GetEnv("DUITKU_RETURN_URL", "")),
	})

	if err != nil {
		return err
	}

	insertPaymentQuery := `
        INSERT INTO payments (
            order_id, price, total_amount, buyer_number, fee,
            fee_amount, status, method, payment_number, created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
        )
    `

	var paymentNumber string
	if duitku.VANumber != "" {
		paymentNumber = duitku.VANumber
	} else if duitku.QrString != "" {
		paymentNumber = duitku.QrString
	} else {
		paymentNumber = duitku.PaymentUrl
	}

	_, err = tx.ExecContext(ctx, insertPaymentQuery,
		orderID,
		fmt.Sprintf("%d", price),
		total,
		whatsApp,
		fee,
		fee,
		"PENDING",
		methodName,
		paymentNumber,
	)

	return err
}

func (repo *TransactionRepository) updateVoucherUsage(ctx context.Context, tx *sql.Tx, voucherCode string) error {
	updateQuery := `
        UPDATE vouchers
        SET usage_count = COALESCE(usage_count, 0) + 1,
            updated_at = NOW()
        WHERE code = $1 AND status = 'active' AND 
              (max_usage IS NULL OR COALESCE(usage_count, 0) < max_usage) AND
              (expires_at IS NULL OR expires_at > NOW())
    `

	result, err := tx.ExecContext(ctx, updateQuery, voucherCode)
	if err != nil {
		return fmt.Errorf("failed to update voucher usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check voucher update result: %w", err)
	}

	if rowsAffected == 0 {
		return ErrVoucherInvalid
	}

	return nil
}

func (repo *TransactionRepository) rollbackOnError(tx *sql.Tx) {
	if rErr := tx.Rollback(); rErr != nil && rErr != sql.ErrTxDone {
		fmt.Printf("Error during transaction rollback: %v\n", rErr)
	}
}

func stringPtr(s string) *string {
	return &s
}
