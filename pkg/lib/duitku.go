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
	QrString      string `json:"qrString,omitempty"`
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
	QrString        string `json:"qrString,omitempty"`
	VANumber        string `json:"vaNumber,omitempty"`
	Amount          string `json:"amount"`
	Reference       string `json:"reference"`
	StatusCode      string `json:"statusCode"`
	StatusMessage   string `json:"statusMessage"`
}

func NewDuitkuService() *DuitkuService {
	return &DuitkuService{
		DuitkuKey:             "9ecc7819ac45c6f63e4351e0329dc123",
		DuitkuMerchantCode:    "D16328",
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

	signature := s.generateSignature(params.MerchantOrderId, params.PaymentAmount)

	payload := map[string]interface{}{
		"merchantCode":    s.DuitkuMerchantCode,
		"paymentAmount":   params.PaymentAmount,
		"merchantOrderId": params.MerchantOrderId,
		"productDetails":  params.ProductDetails,
		"paymentMethod":   params.PaymentCode,
		"signature":       signature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	var duitkuResponse DuitkuCreateTransactionResponse
	if err := json.Unmarshal(body, &duitkuResponse); err != nil {
		return s.createErrorResponse(params.MerchantOrderId, fmt.Sprintf("Failed to parse success response: %v", err)), nil
	}
	return &duitkuResponse, nil
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
