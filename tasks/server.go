package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	zlog "github.com/rs/zerolog/log"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// QueueServer is a task processor that pulls tasks from a queue and execute them
//
// Key Responsibilities:
//
// Manages worker pools: Creates and supervises worker processes (goroutines) that handle task execution.
//
// Dispatches tasks: Assigns tasks from the queues to available worker processes.
//
// Processes tasks: Workers within the server execute the defined task logic using the provided payload.
//
// Handles retries and failures: Manages retries for failed tasks based on configured policies.
type QueueServer struct {
	server *asynq.Server
	store  db.Store
}

// PullNewTask

func PullNewTask(redisOpt asynq.RedisClientOpt, store db.Store) *QueueServer {
	return &QueueServer{
		server: asynq.NewServer(redisOpt, asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
		}),
		store: store,
	}
}

// ProcessVerifyEmail
//
// pull the task from the  redis queue, and feed it to the worker to process the task which is to send the email
func (s *QueueServer) ProcessVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %v", asynq.SkipRetry)
	}
	user, err := s.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist: %v", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// TODO: send email to the user

	zlog.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("email", user.Email).Msg("processed task")

	return nil
}

func (s *QueueServer) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeVerifyEmail, s.ProcessVerifyEmail)
	return s.server.Start(mux)
}
