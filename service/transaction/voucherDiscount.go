package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (repo *TransactionRepository) calculateVoucherDiscount(c context.Context, tx *sql.Tx, voucherCode string, userPrice int) (int, error) {
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

	err := tx.QueryRowContext(c, voucherQuery, voucherCode).Scan(
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
