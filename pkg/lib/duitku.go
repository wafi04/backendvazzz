package lib

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type DuitkuCreateTransactionParams struct {
	PaymentAmount   int     `json:"paymentAmount"`
	MerchantOrderId string  `json:"merchantOrderId"`
	ProductDetails  string  `json:"productDetails"`
	PaymentCode     string  `json:"paymentCode"`
	Cust            *string `json:"cust,omitempty"`
	CallbackUrl     *string `json:"callbackUrl,omitempty"`
	ReturnUrl       *string `json:"returnUrl,omitempty"`
	NoWa            string  `json:"noWa"`
}

type ResponseFromDuitkuCheckTransaction struct {
	Status int `json:"status"`
	Data   struct {
		MerchantOrderId string `json:"merchantOrderId"`
		Reference       string `json:"reference"`
		Amount          string `json:"amount"`
		Fee             string `json:"fee"`
		StatusCode      string `json:"statusCode"`
		StatusMessage   string `json:"statusMessage"`
	} `json:"data"`
}

type DuitkuCreateTransactionResponse struct {
	Status        string      `json:"status"`
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Data          interface{} `json:"data"`
	PaymentUrl    string      `json:"paymentUrl,omitempty"`
	VaNumber      string      `json:"vaNumber,omitempty"`
	Amount        string      `json:"amount,omitempty"`
	Reference     string      `json:"reference,omitempty"`
	StatusCode    string      `json:"statusCode,omitempty"`
	StatusMessage string      `json:"statusMessage,omitempty"`
}

type DuitkuService struct {
	DuitkuKey             string
	DuitkuMerchantCode    string
	DuitkuExpiryPeriod    *int64
	BaseUrl               string
	SandboxUrl            string
	BaseUrlGetTransaction string
	BaseUrlGetBalance     string
	HttpClient            *http.Client
}

func NewDuitkuService(duitkuKey, duitkuMerchantCode string, duitkuExpiryPeriod *int64) *DuitkuService {
	return &DuitkuService{
		DuitkuKey:             duitkuKey,
		DuitkuMerchantCode:    duitkuMerchantCode,
		DuitkuExpiryPeriod:    duitkuExpiryPeriod,
		BaseUrl:               "https://passport.duitku.com/webapi/api/merchant/v2/inquiry",
		SandboxUrl:            "https://sandbox.duitku.com/webapi/api/merchant/v2/inquiry",
		BaseUrlGetTransaction: "https://passport.duitku.com/webapi/api/merchant/transactionStatus",
		BaseUrlGetBalance:     "https://passport.duitku.com/webapi/api/disbursement/checkbalance",
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *DuitkuService) CreateTransaction(ctx context.Context, params *DuitkuCreateTransactionParams) (*DuitkuCreateTransactionResponse, error) {
	// Generate signature using MD5
	signature := s.generateSignature(params.MerchantOrderId, params.PaymentAmount)

	// Prepare payload
	payload := map[string]interface{}{
		"merchantCode":    s.DuitkuMerchantCode,
		"paymentAmount":   params.PaymentAmount,
		"merchantOrderId": params.MerchantOrderId,
		"productDetails":  params.ProductDetails,
		"paymentMethod":   params.PaymentCode,
		"signature":       signature,
		"phoneNumber":     params.NoWa,
	}

	if params.CallbackUrl != nil {
		payload["callbackUrl"] = *params.CallbackUrl
	}
	if params.ReturnUrl != nil {
		payload["returnUrl"] = *params.ReturnUrl
	}
	if s.DuitkuExpiryPeriod != nil {
		payload["expiryPeriod"] = *s.DuitkuExpiryPeriod
	}

	// Make HTTP request
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to marshal request: %v", err)), nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to send request: %v", err)), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to read response: %v", err)), nil
	}

	// Parse response
	var duitkuResponse map[string]interface{}
	if err := json.Unmarshal(body, &duitkuResponse); err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	// Create successful response
	return &DuitkuCreateTransactionResponse{
		Status:  getStringValue(duitkuResponse, "statusMessage"),
		Code:    getStringValue(duitkuResponse, "statusCode"),
		Message: "Transaction created successfully",
		Data: map[string]interface{}{
			"merchantOrderId": params.MerchantOrderId,
			"signature":       signature,
			"timestamp":       time.Now().Format(time.RFC3339),
			"paymentUrl":      getStringValue(duitkuResponse, "paymentUrl"),
			"vaNumber":        getStringValue(duitkuResponse, "vaNumber"),
			"amount":          getStringValue(duitkuResponse, "amount"),
			"reference":       getStringValue(duitkuResponse, "reference"),
			"statusCode":      getStringValue(duitkuResponse, "statusCode"),
			"statusMessage":   getStringValue(duitkuResponse, "statusMessage"),
		},
	}, nil
}

func (s *DuitkuService) GetTransaction(ctx context.Context, merchantOrderId string) (*ResponseFromDuitkuCheckTransaction, error) {
	// Generate signature: md5(merchantCode + merchantOrderId + apiKey)
	signature := s.generateSignature(merchantOrderId, 0)

	payload := map[string]interface{}{
		"merchantcode":    s.DuitkuMerchantCode,
		"merchantOrderId": merchantOrderId,
		"signature":       signature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseUrlGetTransaction, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response ResponseFromDuitkuCheckTransaction
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (s *DuitkuService) GetSaldo(ctx context.Context, email string) (map[string]interface{}, error) {
	timestamp := time.Now().Unix()

	// Generate signature: md5(email + timestamp + apiKey)
	h := md5.New()
	h.Write([]byte(email + strconv.FormatInt(timestamp, 10) + s.DuitkuKey))
	signature := hex.EncodeToString(h.Sum(nil))

	payload := map[string]interface{}{
		"userId":    1,
		"email":     email,
		"timestamp": timestamp,
		"signature": signature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseUrlGetBalance, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

// Helper functions

func (s *DuitkuService) generateSignature(merchantOrderId string, paymentAmount int) string {
	var signatureString string

	if paymentAmount > 0 {
		// For create transaction: merchantCode + merchantOrderId + paymentAmount + apiKey
		signatureString = s.DuitkuMerchantCode + merchantOrderId + strconv.Itoa(paymentAmount) + s.DuitkuKey
	} else {
		// For get transaction: merchantCode + merchantOrderId + apiKey
		signatureString = s.DuitkuMerchantCode + merchantOrderId + s.DuitkuKey
	}

	h := md5.New()
	h.Write([]byte(signatureString))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *DuitkuService) createErrorResponse(merchantOrderId, errorMessage string) *DuitkuCreateTransactionResponse {
	return &DuitkuCreateTransactionResponse{
		Status:  "false",
		Message: errorMessage,
		Data: map[string]interface{}{
			"merchantOrderId": merchantOrderId,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
	}
}

func getStringValue(data map[string]interface{}, key string) string {
	if val, exists := data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
