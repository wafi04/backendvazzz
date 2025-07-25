package transactions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/types"
)

type TransactionsRepository struct {
	DB *sql.DB
}

func NewTransactionsRepository(DB *sql.DB) *TransactionsRepository {
	return &TransactionsRepository{
		DB: DB,
	}
}

func (repo *TransactionsRepository) GetByID(id int64) (*types.Transaction, error) {
	query := `
		SELECT id, order_id, product_code, method_code, game_id, zone, voucher_code,
			   whatsapp_number, nickname, username, ip, user_agent,
			   status, purchase_price, web_price, duitku_price, 
			   created_at, updated_at, completed_at
		FROM transactions WHERE id = $1`

	var transaction types.Transaction
	err := repo.DB.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.OrderId,
		&transaction.ProductCode,
		&transaction.MethodCode,
		&transaction.GameId,
		&transaction.Zone,
		&transaction.VoucherCode,
		&transaction.WhatsAppNumber,
		&transaction.Nickname,
		&transaction.Username,
		&transaction.Ip,
		&transaction.UserAgent,
		&transaction.Status,
		&transaction.PurchasePrice,
		&transaction.WebPrice,
		&transaction.DuitkuPrice,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}
func (repo *TransactionsRepository) GetInvoiceByID(id string) (*model.Invoice, error) {
	query := `
		SELECT
			t.order_id,
			t.username,
			t.discount,
			t.nickname,
			t.user_id,
			t.zone,
			t.message,
			t.serial_number,
			t.status,
			t.created_at,
			COALESCE(p.total_amount, 0) AS total,
			COALESCE(p.status, '') AS paymentStatus,
			COALESCE(p.method, '') AS method,
			COALESCE(p.payment_number, '') AS payementNumber,
			t.updated_at
		FROM transactions t
		LEFT JOIN payments p ON t.order_id = p.order_id
		WHERE t.order_id = $1
	`

	var invoice model.Invoice
	err := repo.DB.QueryRow(query, id).Scan(
		&invoice.OrderID,
		&invoice.Username,
		&invoice.Discount,
		&invoice.Nickname,
		&invoice.UserID,
		&invoice.Zone,
		&invoice.Message,
		&invoice.SerialNumber,
		&invoice.Status,
		&invoice.CreatedAt,
		&invoice.TotalAmount,
		&invoice.PaymentStatus,
		&invoice.Method,
		&invoice.PaymentNumber,
		&invoice.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (repo *TransactionsRepository) GetByGameID(gameId string) ([]types.Transaction, error) {
	query := `
		SELECT   order_id, username, purchase_price, discount, user_id, zone,
            service_name, price, profit, profit_amount, status, is_digi,
            success_report_sent, transaction_type
		FROM transactions WHERE game_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.DB.Query(query, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var transaction types.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.OrderId,
			&transaction.ProductCode,
			&transaction.MethodCode,
			&transaction.GameId,
			&transaction.Zone,
			&transaction.VoucherCode,
			&transaction.WhatsAppNumber,
			&transaction.Nickname,
			&transaction.Username,
			&transaction.Ip,
			&transaction.UserAgent,
			&transaction.Status,
			&transaction.PurchasePrice,
			&transaction.WebPrice,
			&transaction.DuitkuPrice,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (repo *TransactionsRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE transactions 
		SET status = $1, updated_at = $2, completed_at = $3
		WHERE id = $4`

	var completedAt *time.Time
	if status == types.StatusSuccess || status == types.StatusFailed || status == types.StatusCancelled {
		now := time.Now()
		completedAt = &now
	}

	_, err := repo.DB.Exec(query, status, time.Now(), completedAt, id)
	return err
}
func (repo *TransactionsRepository) GetAllWithPayment(c context.Context, req model.FilterTransaction) ([]model.TransactionWithPayment, int, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if req.Search != nil && *req.Search != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf("(t.username ILIKE $%d OR t.order_id ILIKE $%d)", argIndex, argIndex))
		searchPattern := "%" + *req.Search + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	if req.Type != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("t.transaction_type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}

	if req.Status != nil && *req.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.StartDate != nil && *req.StartDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(t.created_at) >= $%d", argIndex))
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil && *req.EndDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(t.created_at) <= $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM transactions t
		LEFT JOIN payments p ON t.order_id = p.order_id
		WHERE %s`, whereClause)

	var totalCount int
	err := repo.DB.QueryRowContext(c, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []model.TransactionWithPayment{}, 0, nil
	}

	offset := (req.Page - 1) * req.Limit

	query := fmt.Sprintf(`
		SELECT 
			t.id, t.order_id, t.username, t.purchase_price, t.discount, t.user_id, t.zone,
			t.nickname, t.service_name, t.price, t.profit, t.message, t.profit_amount,
			t.provider_order_id, t.status, t.log, t.serial_number, t.is_re_order,
			t.transaction_type, t.is_digi, t.ref_id, t.success_report_sent,
			t.created_at, t.updated_at,
			p.order_id as payment_order_id, p.price as payment_price, p.total_amount,
			p.payment_number, p.buyer_number, p.fee, p.fee_amount,
			p.status as payment_status, p.method, p.reference,
			p.created_at as payment_created_at, p.updated_at as payment_updated_at
		FROM transactions t
		LEFT JOIN payments p ON t.order_id = p.order_id
		WHERE %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	rows, err := repo.DB.QueryContext(c, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []model.TransactionWithPayment
	for rows.Next() {
		var transaction model.TransactionWithPayment
		var paymentOrderID, paymentPrice, paymentStatus, method, reference sql.NullString
		var totalAmount, fee, feeAmount sql.NullInt64
		var paymentNumber, buyerNumber sql.NullString
		var paymentCreatedAt, paymentUpdatedAt sql.NullTime

		err := rows.Scan(
			// Transaction fields
			&transaction.ID,
			&transaction.OrderID,
			&transaction.Username,
			&transaction.PurchasePrice,
			&transaction.Discount,
			&transaction.UserID,
			&transaction.Zone,
			&transaction.Nickname,
			&transaction.ServiceName,
			&transaction.Price,
			&transaction.Profit,
			&transaction.Message,
			&transaction.ProfitAmount,
			&transaction.ProviderOrderID,
			&transaction.Status,
			&transaction.Log,
			&transaction.SerialNumber,
			&transaction.IsReOrder,
			&transaction.TransactionType,
			&transaction.IsDigi,
			&transaction.RefID,
			&transaction.SuccessReportSent,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&paymentOrderID,
			&paymentPrice,
			&totalAmount,
			&paymentNumber,
			&buyerNumber,
			&fee,
			&feeAmount,
			&paymentStatus,
			&method,
			&reference,
			&paymentCreatedAt,
			&paymentUpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if paymentOrderID.Valid {
			transaction.PaymentDetail = &model.PaymentDetail{
				OrderID:       paymentOrderID.String,
				Price:         paymentPrice.String,
				TotalAmount:   int(totalAmount.Int64),
				PaymentNumber: paymentNumber.String,
				BuyerNumber:   buyerNumber.String,
				Fee:           int(fee.Int64),
				FeeAmount:     int(feeAmount.Int64),
				Status:        paymentStatus.String,
				Method:        method.String,
				Reference:     reference.String,
				CreatedAt:     paymentCreatedAt.Time,
				UpdatedAt:     paymentUpdatedAt.Time,
			}
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return transactions, totalCount, nil
}
