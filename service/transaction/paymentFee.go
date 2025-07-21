package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wafi04/backendvazzz/pkg/utils"
)

func (repo *TransactionRepository) calculatePaymentFee(c context.Context, tx *sql.Tx, methodCode string, userPrice int) (int, string, error) {
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

	err := tx.QueryRowContext(c, queryFee, methodCode).Scan(&feeValue, &feeType, &methodName)
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
