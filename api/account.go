package api

import (
	"database/sql"
	"net/http"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (s *Server) CreateAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	args := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
	}

	account, err := s.store.CreateAccount(ctx, args)
	if err != nil {
		//TODO: handle error in the request tag validation
		// Check if the error has an SQLState method
		if sqlErr, ok := err.(interface{ SQLState() string }); ok {
			// Access the SQL state code
			sqlState := sqlErr.SQLState()
			//23505  23503
			switch sqlState {
			case "23503", "23505": // Unique violation (might not be accurate)
				httpResponse(ctx, Response{
					Status: http.StatusForbidden,
					Error:  err.Error(),
				})
				return

			}
		}
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  "Internal server error: " + err.Error(),
		})

		return
	}
	data := make(map[string]any)
	data["account"] = account
	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusCreated,
		Error:  "Account created",
	})

}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) GetAccount(ctx *gin.Context) {

	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	account, err := s.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpResponse(ctx, Response{
				Status: http.StatusNotFound,
				Error:  "Account not found",
			})
			return
		}
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}
	data := make(map[string]any)
	data["account"] = account
	httpResponse(ctx, Response{
		Data:     data,
		Status:   http.StatusOK,
		Feedback: "Account created",
	})
}

type listAccountRequest struct {
	Page  int32 `form:"page" binding:"required,min=1"`
	Limit int32 `form:"limit" binding:"required,min=5,max=25"`
}

func (s *Server) ListAccounts(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	args := db.ListAccountsParams{
		Offset: (req.Page - 1) * req.Limit,
		Limit:  req.Limit,
	}

	accounts, err := s.store.ListAccounts(ctx, args)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})

		return
	}
	data := make(map[string]any)
	data["accounts"] = accounts
	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusOK,
	})
}
