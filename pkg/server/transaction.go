package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/utils"
	"github.com/wafi04/backendvazzz/service/transaction"
	"github.com/wafi04/backendvazzz/service/transactions"
)

type RequestFromClient struct {
	ProductCode string  `json:"productCode" validate:"required"`
	MethodCode  string  `json:"methodCode" validate:"required"`
	WhatsApp    string  `json:"whatsapp" validate:"required"`
	VoucherCode *string `json:"voucherCode,omitempty"`
	GameId      string  `json:"gameId" validate:"required"`
	Zone        *string `json:"zone,omitempty"`
}

func StringPtr(s string) *string {
	return &s
}
func SetUpTransactionRoutes(api *gin.RouterGroup, db *sql.DB) {
	transactionRepo := transaction.NewTransactionRepository(db)
	transactionsRepo := transactions.NewTransactionsRepository(db)
	transactionsHandler := transactions.NewTransactionHandler(transactionsRepo)

	r := api.Group("/transactions")
	{

		r.POST("", func(ctx *gin.Context) {
			var input RequestFromClient
			if err := ctx.ShouldBindJSON(&input); err != nil {
				utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid input", err.Error())
				return
			}

			// Validasi input
			if input.ProductCode == "" {
				utils.ErrorResponse(ctx, http.StatusBadRequest, "Product code is required", "")
				return
			}
			if input.MethodCode == "" {
				utils.ErrorResponse(ctx, http.StatusBadRequest, "Method code is required", "")
				return
			}
			if input.WhatsApp == "" {
				utils.ErrorResponse(ctx, http.StatusBadRequest, "WhatsApp number is required", "")
				return
			}
			if input.GameId == "" {
				utils.ErrorResponse(ctx, http.StatusBadRequest, "Game ID is required", "")
				return
			}

			var usernamePtr *string
			if u, ok := ctx.Get("username"); ok {
				if usernameStr, isString := u.(string); isString {
					usernamePtr = StringPtr(usernameStr)
				}
			}

			// fmt.Printf("%s", *usernamePtr)

			var rolePtr *string
			if r, ok := ctx.Get("role"); ok {
				if roleStr, isString := r.(string); isString {
					rolePtr = StringPtr(roleStr)
				}
			}

			response, err := transactionRepo.Create(ctx, transaction.CreateTransaction{
				ProductCode: input.ProductCode,
				MethodCode:  input.MethodCode,
				WhatsApp:    input.WhatsApp,
				Username:    usernamePtr,
				Role:        rolePtr,
				VoucherCode: input.VoucherCode,
				GameId:      input.GameId,
				Zone:        input.Zone,
			})
			if err != nil {
				utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create transaction", err.Error())
				return
			}

			utils.SuccessResponse(ctx, http.StatusCreated, "Transaction created successfully", response)
		})

		r.GET("", transactionsHandler.GetAll)
		r.GET("/invoice/:id", transactionsHandler.Invoice)
	}

}
