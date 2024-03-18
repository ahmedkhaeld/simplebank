package gapi

import (
	"context"
	"errors"
	"time"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/pb"
	"github.com/ahmedkhaeld/simplebank/tasks"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/ahmedkhaeld/simplebank/val"
	"github.com/hibiken/asynq"
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

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username: req.GetUsername(),
			Password: hashPassword,
			FullName: req.GetFullName(),
			Email:    req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &tasks.PayloadVerifyEmail{
				Username: user.Username,
			}
			//define queue options to be in the critical queue, with 10 max retries,
			// process the task after 10 seconds so the worker when pick up the task, which is dependent on on  the users table
			//worker must found a user row for in the db table
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(tasks.QueueCritical),
			}
			return server.taskClient.VerifyEmail(ctx, taskPayload, opts...)

		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)
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
			Username:          txResult.User.Username,
			FullName:          txResult.User.FullName,
			Email:             txResult.User.Email,
			CreatedAt:         timestamppb.New(txResult.User.CreatedAt),
			PasswordChangedAt: timestamppb.New(txResult.User.PasswordChangedAt),
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
