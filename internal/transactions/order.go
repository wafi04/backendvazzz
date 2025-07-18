package transactions

import (
	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/types"
)

func (repo *TransactionsRepository) CreateOrder(create *types.CreateTransactions) {

	digi := lib.NewDigiflazzService(lib.DigiConfig{
		DigiKey:      config.GetEnv("DIGI_KEY", ""),
		DigiUsername: config.GetEnv("DIGI_USERNAME", ""),
	})

}
