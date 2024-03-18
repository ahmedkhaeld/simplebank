package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	zlog "github.com/rs/zerolog/log"

	"github.com/hibiken/asynq"
)

const TypeVerifyEmail = "email: verification-email"

// QueueClient is a task submitter
//
// Key Responsibilities:
//
// Defines tasks: Specifies the data to be processed (payload) and the type of task to be executed.
//
// Enqueues tasks: Pushes tasks onto the designated queue(s) managed by the Asynq server.
//
// Optional configuration: Can set options like retries, priority, or delays for specific tasks.
type QueueClient struct {
	client *asynq.Client
}

// PushNewTask Pushes tasks onto the designated queue(s) managed by the Asynq server.
func PushNewTask(redisOpt asynq.RedisClientOpt) QueueClient {
	return QueueClient{
		client: asynq.NewClient(redisOpt),
	}
}

type PayloadVerifyEmail struct {
	Username string `json:"username`
}

func (c *QueueClient) VerifyEmail(ctx context.Context, payload *PayloadVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %v", err)
	}

	task := asynq.NewTask(TypeVerifyEmail, jsonPayload, opts...)
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %v", err)
	}
	zlog.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}
