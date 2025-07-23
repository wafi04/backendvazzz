package transactions

import (
	"context"
	"database/sql"
)

type CallbackData struct {
	Data CallbackDetail `json:"data"`
}

type CallbackDetail struct {
	RefID        string `json:"ref_id"`
	BuyerSKUCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	Status       string `json:"status"` // bisa jadi enum kalau perlu
	Message      string `json:"message"`
	SN           string `json:"sn"` // serial number, bisa juga diganti jadi `SerialNumber` jika ingin lebih jelas
}

type CallbackDigi struct {
	DB *sql.DB
}

func NewCallbackDigi(db *sql.DB) *CallbackDigi {
	return &CallbackDigi{
		DB: db,
	}
}

func (db *CallbackDigi) Callback(c context.Context, data CallbackData) error {
	Data := CallbackData.Data
	Data,err != nil {

	}


	
}
