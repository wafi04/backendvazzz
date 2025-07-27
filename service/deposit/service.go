package deposit

import (
	"context"
	"fmt"
	"log"

	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type DepositService struct {
	repo *DepositRepository
}

func NewDepositService(repo *DepositRepository) *DepositService {
	return &DepositService{
		repo: repo,
	}
}

func stringPtr(s string) *string {
	return &s
}

func (service *DepositService) Create(c context.Context, amount int, methodCode, username string) (string, error) {
	duitku := lib.NewDuitkuService()
	depStr := "DEP"
	depositID := utils.GenerateUniqeID(&depStr)

	duitkuCall, err := duitku.CreateTransaction(c, &lib.DuitkuCreateTransactionParams{
		PaymentAmount:   amount,
		MerchantOrderId: depositID,
		ProductDetails:  "Deposit",
		PaymentCode:     methodCode,
		CallbackUrl:     stringPtr(config.GetEnv("DUITKU_CALLBACK_URL", "")),
		ReturnUrl:       stringPtr(config.GetEnv("DUITKU_RETURN_URL", "")),
	})

	log.Println("Duitku response:", duitkuCall)
	if err != nil {
		log.Printf("Error calling Duitku API: %v", err)
		return "", fmt.Errorf("failed to create payment transaction: %w", err) // Fix: return actual error
	}

	// Create deposit record in database
	_, err = service.repo.Create(c, model.CreateDeposit{
		Method:           methodCode,
		Amount:           amount,
		PaymentReference: duitkuCall.PaymentUrl,
	}, username, depositID, "Deposit Pending")

	if err != nil {
		log.Printf("Error saving deposit to database: %v", err)
		return "", fmt.Errorf("failed to save deposit: %w", err) // Fix: return actual error
	}

	return depositID, nil
}

func (serv *DepositService) GetAll(c context.Context, limit, offset int, search, status string) ([]model.DepositData, int, error) {
	return serv.repo.GetAll(c, limit, offset, search, status)
}

func (serv *DepositService) GetAllByUsername(c context.Context, limit, offset int, username string) ([]model.DepositData, int, error) {
	return serv.repo.GetByDepositByUsername(c, username, limit, offset)
}

func (serv *DepositService) GetByDepositID(c context.Context, depositID string) (*model.DepositData, error) {
	return serv.repo.GetByDepositID(c, depositID)
}

func (serv *DepositService) Delete(c context.Context, id int) error {
	return serv.repo.Delete(c, id)
}
