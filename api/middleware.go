package api

import (
	"errors"
	"strings"

	"github.com/ahmedkhaeld/simplebank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

var (
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
	ErrInvalidAuthorizationType   = errors.New("invalid authorization type")
	ErrInvalidAuthorizationFormat = errors.New("invalid authorization format")
)

// BearerMiddleware creates a custom gin middleware for authentication with Bearer
func BearerMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			httpUnauthorized(ctx, ErrInvalidAuthorizationHeader)
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			httpUnauthorized(ctx, ErrInvalidAuthorizationFormat)
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			httpUnauthorized(ctx, ErrInvalidAuthorizationType)
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			httpUnauthorized(ctx, err)
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
