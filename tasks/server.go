package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	zlog "github.com/rs/zerolog/log"

	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/mail"
	"github.com/ahmedkhaeld/simplebank/util"
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
	mailer mail.Mailer
}

// PullNewTask

func PullNewTask(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.Mailer) *QueueServer {
	return &QueueServer{
		server: asynq.NewServer(redisOpt, asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				zlog.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: NewLogger(),
		}),
		store:  store,
		mailer: mailer,
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
		// Note: ignore the skipRetry error
		//create user tx will eventually be created, no need to do an error check if not found when the worker is created before this
		// make the retry function do it thing, to try again to process the task
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user does not exist: %v", asynq.SkipRetry)
		// }
		return fmt.Errorf("failed to find user: %w", err)
	}

	// create verify_email record and  send email to the user
	verifyemail, err := s.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	//mail parameters
	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:8080/api/v1/verify_email?email_id=%d&secret_code=%s",
		verifyemail.ID, verifyemail.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.FullName, verifyUrl)
	to := []string{user.Email}

	err = s.mailer.Send(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	zlog.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("email", user.Email).Msg("processed task")

	return nil
}

func (s *QueueServer) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeVerifyEmail, s.ProcessVerifyEmail)
	return s.server.Start(mux)
}
