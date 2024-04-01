package gapi

import (
	"context"
	"strings"

	"github.com/ahmedkhaeld/simplebank/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authenticate(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not found")
	}
	values := md.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header not found")
	}
	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect authorization header format")
	}
	authType := strings.ToLower(fields[0])
	if authType != authorizationTypeBearer {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect authorization header type")
	}
	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect access token")
	}
	return payload, nil
}
