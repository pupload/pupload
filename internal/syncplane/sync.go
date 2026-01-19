package syncplane

import (
	"context"
	"time"
)

type SyncLayer interface {
	RegisterExecuteNodeHandler(handler ExecuteNodeHandler) error
	EnqueueExecuteNode(payload NodeExecutePayload) error

	RegisterNodeFinishedHandler(handler NodeFinishedHandler) error
	EnqueueNodeFinished(payload NodeFinishedPayload) error

	RegisterNodeErrorHandler(handler NodeErrorHandler) error
	EnqueueNodeError(payload NodeErrorHandler) error

	UpdateSubscribedQueues(queues map[string]int) error

	RegisterFlowStepHandler(handler FlowStepHandler) error
	StartScheduler(ctx context.Context)
	StopScheduler(ctx context.Context)
	AddRunToScheduler(run_id string) error
	RemoveRunFromScheduler(run_id string) error

	NewMutex(run_id string, duration time.Duration) Mutex

	Start() error
	Close() error
}

type Mutex interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}

type Task struct {
	Task    string
	Payload []byte
}

type SyncPlaneSettings struct {
	SelectedSyncPlane string

	Redis RedisSettings

	ControllerStepInterval string // time inbetween a flowruns attemped steps, written as cronspec. eg. @every 10s
}

type RedisSettings struct {
	Address  string
	Password string
	DB       int

	PoolSize   int
	MaxRetries int

	// TODO: add redis TLS support
}
