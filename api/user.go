package api

import (
	"net/http"
	"time"

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

// userResponse purposes execlude the password from the response body
type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
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
		//TODO: handle error in the request tag validation
		if sqlErr, ok := err.(interface{ SQLState() string }); ok {
			sqlState := sqlErr.SQLState()
			switch sqlState {
			case "23503", "23505":
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
	data["user"] = userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	httpResponse(ctx, Response{
		Data:   data,
		Status: http.StatusCreated,
	})

}
