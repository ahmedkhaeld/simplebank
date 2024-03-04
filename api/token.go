package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusUnauthorized,
			Error:  err.Error(),
		})
		return
	}

	parsedUUID, err := uuid.Parse(refreshPayload.MapClaims["iss"].(string))
	if err != nil {
		fmt.Println("Error parsing UUID:", err)
		return
	}

	session, err := server.store.GetSession(ctx, parsedUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpResponse(ctx, Response{
				Status: http.StatusNotFound,
				Error:  "Session not found",
			})
			return
		}
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  "Internal server error: " + err.Error(),
		})
		return
	}

	//check if session is blocked or not
	if session.IsBlocked {
		httpResponse(ctx, Response{
			Status: http.StatusUnauthorized,
			Error:  "Session is blocked",
		})
		return
	}

	// check the session username
	if session.Username != refreshPayload.MapClaims["sub"] {
		httpResponse(ctx, Response{
			Status: http.StatusUnauthorized,
			Error:  "Invalid user",
		})
		return
	}

	if session.RefreshToken != req.RefreshToken {
		httpResponse(ctx, Response{
			Status: http.StatusUnauthorized,
			Error:  "Invalid refresh token",
		})
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.MapClaims["sub"].(string), server.env.RefreshTokenDuration)
	if err != nil {
		httpResponse(ctx, Response{
			Status: http.StatusInternalServerError,
			Error:  "Internal server error: " + err.Error(),
		})
		return
	}

	accessPayloadExpire, _ := accessPayload.MapClaims["exp"].(int64)

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: time.Unix(accessPayloadExpire, 0),
	}

	httpResponse(ctx, Response{
		Data:   rsp,
		Status: http.StatusOK,
	})

}
