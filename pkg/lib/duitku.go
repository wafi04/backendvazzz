package lib

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
}

type Duitku struct {
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
	Status        string `json:"status"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	QRCode        string `json:"qrCode,omitempty"`
	VANumber      string `json:"vaNumber,omitempty"`
	PaymentUrl    string `json:"paymentUrl,omitempty"`
	Amount        string `json:"amount,omitempty"`
	Reference     string `json:"reference,omitempty"`
	StatusCode    string `json:"statusCode,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
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

type PaymentResponse struct {
	MerchantOrderId string `json:"merchantOrderId"`
	Signature       string `json:"signature"`
	Timestamp       string `json:"timestamp"`
	PaymentUrl      string `json:"paymentUrl"`
	QRCode          string `json:"qrCode,omitempty"`
	VANumber        string `json:"vaNumber,omitempty"`
	Amount          string `json:"amount"`
	Reference       string `json:"reference"`
	StatusCode      string `json:"statusCode"`
	StatusMessage   string `json:"statusMessage"`
}

func NewDuitkuService() *DuitkuService {
	return &DuitkuService{
		// DuitkuKey:             "D16328",
		// DuitkuMerchantCode:    "9ecc7819ac45c6f63e4351e0329dc123",
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
	log.Printf("INFO: Attempting to create Duitku transaction for order: %s\n", params.MerchantOrderId)

	signature := s.generateSignature(params.MerchantOrderId, params.PaymentAmount)
	log.Printf("DEBUG: Generated signature: %s\n", signature)

	payload := map[string]interface{}{
		"merchantCode":    s.DuitkuMerchantCode,
		"paymentAmount":   params.PaymentAmount,
		"merchantOrderId": params.MerchantOrderId,
		"productDetails":  params.ProductDetails,
		"paymentMethod":   params.PaymentCode,
		"signature":       signature,
	}

	log.Printf("DEBUG: Duitku Request Payload: %+v\n", payload)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: Failed to marshal request for order %s: %v\n", params.MerchantOrderId, err)
		return nil, nil
	}
	log.Printf("DEBUG: Duitku Request JSON: %s\n", string(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ERROR: Failed to create request for order %s: %v\n", params.MerchantOrderId, err)
		return nil, nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: Failed to send request for order %s: %v\n", params.MerchantOrderId, err)
		return nil, nil
	}
	defer resp.Body.Close()

	log.Printf("INFO: Duitku Response Status for order %s: %s\n", params.MerchantOrderId, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response body for order %s: %v\n", params.MerchantOrderId, err)
		return nil, nil
	}
	log.Printf("DEBUG: Duitku Raw Response Body for order %s: %s\n", params.MerchantOrderId, string(body))

	var duitkuResponse DuitkuCreateTransactionResponse
	if err := json.Unmarshal(body, &duitkuResponse); err != nil {
		log.Printf("ERROR: Failed to parse SUCCESS response for order %s: %v. Raw body: %s\n", params.MerchantOrderId, err, string(body))
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to parse success response: %v", err)), nil
	}
	log.Printf("INFO: Duitku Parsed Success Response for order %s: %+v\n", params.MerchantOrderId, duitkuResponse)
	return &duitkuResponse, nil
}

func (s *DuitkuService) GetSaldo(ctx context.Context, email string) (map[string]interface{}, error) {
	timestamp := time.Now().Unix()

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

func (s *DuitkuService) generateSignature(merchantOrderId string, paymentAmount int) string {

	signatureString := s.DuitkuMerchantCode + merchantOrderId + strconv.Itoa(paymentAmount) + s.DuitkuKey

	h := md5.New()
	h.Write([]byte(signatureString))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *DuitkuService) createErrorResponse(merchantOrderId, errorMessage string) *DuitkuCreateTransactionResponse {
	return &DuitkuCreateTransactionResponse{
		Status:  "false",
		Message: errorMessage,
	}
}
