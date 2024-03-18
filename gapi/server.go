package gapi

import (
	"fmt"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/pb"
	"github.com/ahmedkhaeld/simplebank/tasks"
	"github.com/ahmedkhaeld/simplebank/token"
	"github.com/ahmedkhaeld/simplebank/util"
)

// Server serves gRPC requests for our banking service.
type Server struct {
	pb.UnimplementedSimpleBankServer
	env        util.Env
	store      db.Store
	tokenMaker token.Maker
	taskClient tasks.QueueClient
}

// NewServer creates a new gRPC server.
func NewServer(env util.Env, store db.Store, taskClient tasks.QueueClient) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(env.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		env:        env,
		store:      store,
		tokenMaker: tokenMaker,
		taskClient: taskClient,
	}

	return server, nil
}
