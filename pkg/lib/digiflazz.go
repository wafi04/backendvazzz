package lib

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/types"
)

type DigiConfig struct {
	DigiKey      string
	DigiUsername string
}

func NewDigiflazzService(digiConfig DigiConfig) *DigiConfig {
	return &DigiConfig{
		DigiKey:      digiConfig.DigiKey,
		DigiUsername: digiConfig.DigiUsername,
	}
}

func (digi *DigiConfig) Topup(create types.CreateTopup) (*types.ResponseFromDigiflazz, error) {
	var noTujuan string

	// Create signature
	strs := []string{digi.DigiUsername, digi.DigiKey, create.Reference}
	joined := strings.Join(strs, "")
	data := []byte(joined)
	hash := sha256.Sum256(data)
	signature := hex.EncodeToString(hash[:])

	if create.ServerId == "" {
		noTujuan = create.UserId
	} else {
		noTujuan = create.UserId + create.ServerId
	}

	// Prepare request data
	requestData := types.DigiflazzRequest{
		Username:     digi.DigiUsername,
		BuyerSkuCode: create.ProductCode,
		CustomerNo:   noTujuan,
		RefId:        create.Reference,
		Sign:         signature,
		CallbackURL:  config.GetEnv("DIGIFLAZZ_CALLBACK_URL", ""),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.digiflazz.com/v1/transaction", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var result types.ResponseFromDigiflazz
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &result, nil
}
