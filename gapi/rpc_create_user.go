package gapi

import (
	"context"
	"errors"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/pb"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/ahmedkhaeld/simplebank/val"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	violatoins := validateCreateUserRequest(req)
	if violatoins != nil {
		return nil, invalidArgumentError(violatoins)
	}

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

func validateCreateUserRequest(req *pb.CreateUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violation = append(violation, fieldViolation("email", err))
	}
	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violation = append(violation, fieldViolation("full_name", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, fieldViolation("password", err))
	}

	return violation

}
