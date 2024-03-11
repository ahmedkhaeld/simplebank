package gapi

import (
	"context"
	"errors"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/pb"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hashPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}
	arg := db.CreateUserParams{
		Username: req.GetUsername(),
		Password: hashPassword,
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, status.Errorf(codes.AlreadyExists, "username or email already exists")
			}

			return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
		}
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			CreatedAt:         timestamppb.New(user.CreatedAt),
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		},
	}, nil

}
