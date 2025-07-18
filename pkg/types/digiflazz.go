package types

type CreateTopup struct {
	UserId      string
	ServerId    string
	Reference   string
	ProductCode string
}

// Request structure for Digiflazz API
type DigiflazzRequest struct {
	Username     string `json:"username"`
	BuyerSkuCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	RefId        string `json:"ref_id"`
	Sign         string `json:"sign"`
	CallbackURL  string `json:"cb_url"`
}

type ResponseFromDigiflazz struct {
	Data struct {
		RefId          string `json:"ref_id"`
		Status         string `json:"status"`
		Message        string `json:"message"`
		BuyerSkuCode   string `json:"buyer_sku_code"`
		BuyerLastSaldo int    `json:"buyer_last_saldo"`
		Price          int    `json:"price"`
		TrxId          string `json:"trx_id"`
		Rc             string `json:"rc"`
		Sn             string `json:"sn"`
	} `json:"data"`
}
