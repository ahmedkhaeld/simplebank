package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) CreateTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	if !s.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !s.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	args := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := s.store.TransferTx(ctx, args)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}
	data := make(map[string]any)
	data["transfer"] = result
	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusCreated,
		Error:  "Money Transfer Successfully",
	})

}

func (s *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {

	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpResponse(ctx, Response{
				Status: http.StatusNotFound,
				Error:  err.Error(),
			})
			return false
		}
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
	}

	if account.Currency != currency {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  fmt.Sprintf("Account %d does not have currency %s", accountID, currency),
		})
		return false
	}

	return true
}
