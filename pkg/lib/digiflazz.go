package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ProductData represents individual product data
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

// DigiflazzResponse represents the main API response structure
type DigiflazzResponse struct {
	Data []ProductData `json:"data"` // Array of products
}

// DigiConfig holds Digiflazz configuration
type DigiConfig struct {
	DigiKey      string
	DigiUsername string
}

// DigiflazzService handles Digiflazz API operations
type DigiflazzService struct {
	config DigiConfig
}

// NewDigiflazzService creates new Digiflazz service instance
func NewDigiflazzService(config DigiConfig) *DigiflazzService {
	return &DigiflazzService{
		config: config,
	}
}

// generateSign creates MD5 signature for API request
func (d *DigiflazzService) generateSign(username, apiKey, cmd string) string {
	data := username + apiKey + cmd
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// CheckPrice fetches price list from Digiflazz API
func (d *DigiflazzService) CheckPrice() ([]*ProductData, error) {
	// Prepare request payload
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

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Try to unmarshal as DigiflazzResponse first (with data wrapper)
	var apiResponse DigiflazzResponse
	if err := json.Unmarshal(body, &apiResponse); err == nil && len(apiResponse.Data) > 0 {
		// Convert to pointer slice
		result := make([]*ProductData, len(apiResponse.Data))
		for i := range apiResponse.Data {
			result[i] = &apiResponse.Data[i]
		}
		return result, nil
	}

	// If that fails, try to unmarshal as direct array
	var directArray []ProductData
	if err := json.Unmarshal(body, &directArray); err == nil {
		// Convert to pointer slice
		result := make([]*ProductData, len(directArray))
		for i := range directArray {
			result[i] = &directArray[i]
		}
		return result, nil
	}

	// If both fail, return error with response body for debugging
	return nil, fmt.Errorf("failed to unmarshal response. Response body: %s", string(body))
}
