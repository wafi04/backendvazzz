package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wafi04/backendvazzz/pkg/utils"
	"github.com/wafi04/backendvazzz/service/transaction"
	"github.com/wafi04/backendvazzz/service/transactions"
)

func SetUpTransactionRoutes(api *gin.RouterGroup, db *sql.DB) {
	transactionRepo := transaction.NewTransactionRepository(db)
	transactionsRepo := transactions.NewTransactionsRepository(db)
	transactionsHandler := transactions.NewTransactionHandler(transactionsRepo)

	r := api.Group("/transactions")
	{
		r.POST("", func(ctx *gin.Context) {
			var input transaction.CreateTransaction
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

			response, err := transactionRepo.Create(ctx, input)
			if err != nil {
				utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create transaction", err.Error())
				return
			}

			utils.SuccessResponse(ctx, http.StatusCreated, "Transaction created successfully", response)
		})

		r.GET("", transactionsHandler.GetAll)
	}

}
