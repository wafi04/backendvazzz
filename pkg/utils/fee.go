package utils

import (
	"fmt"
	"math"
	"time"
)

func CalculateFeeQris(amount int) int {
	result := float64(amount) * (0.7 / 100)
	calculatedFee := int(math.Ceil(result))
	return calculatedFee
}

func GenerateUniqeID(uniqe *string) string {
	var orderID string
	if uniqe != nil {
		orderID = fmt.Sprintf("%s%d", *uniqe, time.Now().Unix())
		return orderID
	}
	orderID = fmt.Sprintf("TRX%d", time.Now().Unix())
	return orderID
}
