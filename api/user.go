package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (server *Server) CreateUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		Password: hashedPassword,
		FullName: req.FullName,
		Email:    req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		//TODO: handle error in the request tag validation, prevent user to enter existing username or email
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				httpResponse(
					ctx,
					Response{
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
	data["user"] = user

	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusCreated,
	})

}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) LoginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			httpResponse(ctx, Response{
				Status: http.StatusNotFound,
				Error:  "User not found",
			})
			return
		}
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  "Internal server error: " + err.Error(),
		})
		return
	}

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusUnauthorized,
			Error:  err.Error(),
		})
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.env.AccessTokenDuration)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  "Internal server error: " + err.Error(),
		})
		return
	}

	data := make(map[string]any)
	data["user"] = user
	data["access_token"] = accessToken

	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusOK,
	})

}
