package model

import "time"

type Services struct {
	ID               int        `json:"id"`
	CategoryID       int        `json:"category_id"`
	SubCategoryID    int        `json:"sub_category_id"`
	ProviderID       string     `json:"provider_id"`
	ServiceName      string     `json:"service_name"`
	Price            int        `json:"price"`
	PricePurchase    int        `json:"price_purchase"`
	PriceReseller    int        `json:"price_reseller"`
	PricePlatinum    int        `json:"price_platinum"`
	PriceFlashSale   *int       `json:"price_flash_sale"`
	PriceSuggest     *int       `json:"price_suggest"`
	Profit           int        `json:"profit"`
	ProfitReseller   int        `json:"profit_reseller"`
	ProfitPlatinum   int        `json:"profit_platinum"`
	ProfitSuggest    *int       `json:"profit_suggest"`
	IsProfitFixed    string     `json:"is_profit_fixed"`
	IsFlashSale      string     `json:"is_flash_sale"`
	IsSuggest        string     `json:"is_suggest"`
	TitleFlashSale   *string    `json:"title_flash_sale"`
	BannerFlashSale  *string    `json:"banner_flash_sale"`
	ExpiredFlashSale *time.Time `json:"expired_flash_sale"`
	Note             string     `json:"note"`
	Status           string     `json:"status"`       // default 'active'
	Provider         string     `json:"provider"`     // default 'DIGIFLAZZ'
	ProductLogo      *string    `json:"product_logo"` // TEXT, NULLABLE
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
