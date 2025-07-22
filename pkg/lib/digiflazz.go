package lib

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProductData struct {
	BuyerProductStatus  bool   `json:"buyer_product_status"`
	BuyerSkuCode        string `json:"buyer_sku_code"`
	Category            string `json:"category"`
	Desc                string `json:"desc"`
	EndCutOff           string `json:"end_cut_off"`
	Multi               bool   `json:"multi"`
	Price               int    `json:"price"`
	ProductName         string `json:"product_name"`
	SellerName          string `json:"seller_name"`
	SellerProductStatus bool   `json:"seller_product_status"`
	StartCutOff         string `json:"start_cut_off"`
	Stock               int    `json:"stock"`
	Type                string `json:"type"`
	UnlimitedStock      bool   `json:"unlimited_stock"`
	Brand               string `json:"brand"`
}

type CreateTransactionToDigiflazz struct {
	Username     string `json:"username"`
	BuyerSKUCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	RefID        string `json:"ref_id"`
	CallbackURL  string `json:"cb_url,omitempty"`
}
type TransactionCreateDigiflazzResponse struct {
	Data struct {
		RefID          string `json:"ref_id"`           // ID unik dari transaksi
		CustomerNo     string `json:"customer_no"`      // Nomor pelanggan
		BuyerSKUCode   string `json:"buyer_sku_code"`   // Kode produk
		Message        string `json:"message"`          // Pesan status transaksi
		Status         string `json:"status"`           // Status transaksi, contoh: Pending
		RC             string `json:"rc"`               // Response code
		SN             string `json:"sn"`               // Serial number (bisa kosong)
		BuyerLastSaldo int    `json:"buyer_last_saldo"` // Saldo terakhir pembeli
		Price          int    `json:"price"`            // Harga transaksi
		Tele           string `json:"tele"`             // Kontak Telegram
		WA             string `json:"wa"`               // Kontak WhatsApp
	} `json:"data"`
}

type DigiflazzErrorResponse struct {
	Data struct {
		Message string `json:"message"`
		RC      string `json:"rc"`
	} `json:"data"`
}

type DigiflazzResponse struct {
	Data []ProductData `json:"data"`
}

type DigiConfig struct {
	DigiKey      string
	DigiUsername string
}

type DigiflazzService struct {
	config DigiConfig
}

func NewDigiflazzService(config DigiConfig) *DigiflazzService {
	return &DigiflazzService{
		config: config,
	}
}

func (d *DigiflazzService) generateSign(username, apiKey, cmd string) string {
	data := username + apiKey + cmd
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (d *DigiflazzService) CheckPrice() ([]*ProductData, error) {
	sign := d.generateSign(d.config.DigiUsername, d.config.DigiKey, "pricelist")

	requestPayload := map[string]interface{}{
		"username": d.config.DigiUsername,
		"cmd":      "pricelist",
		"sign":     sign,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make HTTP request
	resp, err := http.Post("https://api.digiflazz.com/v1/price-list", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResponse DigiflazzResponse
	if err := json.Unmarshal(body, &apiResponse); err == nil && len(apiResponse.Data) > 0 {
		result := make([]*ProductData, len(apiResponse.Data))
		for i := range apiResponse.Data {
			result[i] = &apiResponse.Data[i]
		}
		return result, nil
	}

	var directArray []ProductData
	if err := json.Unmarshal(body, &directArray); err == nil {
		result := make([]*ProductData, len(directArray))
		for i := range directArray {
			result[i] = &directArray[i]
		}
		return result, nil
	}

	return nil, fmt.Errorf("failed to unmarshal response. Response body: %s", string(body))
}

func (d *DigiflazzService) TopUp(ctx context.Context, req CreateTransactionToDigiflazz) (*TransactionCreateDigiflazzResponse, error) {
	// Generate signature
	data := d.config.DigiUsername + d.config.DigiKey + req.RefID
	hash := md5.Sum([]byte(data))
	sign := fmt.Sprintf("%x", hash)

	requestPayload := map[string]interface{}{
		"username":       d.config.DigiUsername,
		"buyer_sku_code": req.BuyerSKUCode,
		"customer_no":    req.CustomerNo,
		"ref_id":         req.RefID,
		"sign":           sign,
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.digiflazz.com/v1/transaction", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "DigiflazzClient/1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse TransactionCreateDigiflazzResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	switch apiResponse.Data.Status {
	case "Sukses":
		return &apiResponse, nil
	case "Pending":
		return &apiResponse, nil
	case "Gagal":
		return &apiResponse, nil
	default:
		return &apiResponse, nil
	}

}
