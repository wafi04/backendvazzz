package utils

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	counter     uint32
	counterLock sync.Mutex
	lastNano    int64
)

func CalculateFeeQris(amount int) int {
	result := float64(amount) * (0.7 / 100)
	calculatedFee := int(math.Ceil(result))
	return calculatedFee
}

func GenerateUniqeID(prefix *string) string {
	counterLock.Lock()
	defer counterLock.Unlock()

	nowNano := time.Now().UnixNano()

	if nowNano == lastNano {
		counter++
	} else {
		counter = 0
		lastNano = nowNano
	}

	var orderID string
	if prefix != nil && *prefix != "" {
		orderID = fmt.Sprintf("%s%d%04d", *prefix, nowNano, counter)
	} else {
		orderID = fmt.Sprintf("TRX%d%04d", nowNano, counter)
	}

	return orderID
}
