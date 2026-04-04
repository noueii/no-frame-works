package provider

import (
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type AsynqClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}

type AsynqProvider struct {
	Client AsynqClient
}

type MockAsynqClient struct {
	Tasks []asynq.Task
}

func (p *MockAsynqClient) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	p.Tasks = append(p.Tasks, *task)

	return &asynq.TaskInfo{
		ID:    "mock-task-" + uuid.New().String(),
		Queue: "mock-queue",
	}, nil
}

func (p *MockAsynqClient) Close() error {
	return nil
}

func NewQueueProvider(rdb *redis.Client) *AsynqProvider {
	opts := rdb.Options()

	return &AsynqProvider{
		Client: asynq.NewClient(asynq.RedisClientOpt{
			Addr:      opts.Addr,
			Password:  opts.Password,
			DB:        opts.DB,
			TLSConfig: opts.TLSConfig,
		}),
	}
}

func NewTestQueueProvider(_ *EnvProvider) *AsynqProvider {
	return &AsynqProvider{
		Client: &MockAsynqClient{},
	}
}
