package syncplane

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/resources"

	"github.com/cusianovic/asynq"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type RedisSync struct {
	redisClient *redis.Client

	asynqClient *asynq.Client
	asynqServer *asynq.Server
	redsync     *redsync.Redsync

	scheduler *asynq.PeriodicTaskManager

	mux *asynq.ServeMux

	workerResourceManger *resources.ResourceManager

	mu sync.Mutex

	log *slog.Logger
}

func NewControllerRedisSyncLayer(cfg SyncPlaneSettings) *RedisSync {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,

		MaxRetries: cfg.Redis.MaxRetries,
		PoolSize:   cfg.Redis.PoolSize,
	})

	asynqClient := asynq.NewClientFromRedisClient(rdb)
	asynqServer := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"controller": 1,
		},

		LogLevel: asynq.FatalLevel,
	})

	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	redisSync := &RedisSync{
		redisClient: rdb,
		asynqClient: asynqClient,
		asynqServer: asynqServer,
		redsync:     rs,

		mux: asynq.NewServeMux(),

		log: logging.ForService("controller-synclayer"),
	}

	mgr, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		RedisUniversalClient:       rdb,
		PeriodicTaskConfigProvider: newSchedulerTaskProvider(redisSync, cfg.ControllerStepInterval),
		SyncInterval:               10 * time.Second,

		SchedulerOpts: &asynq.SchedulerOpts{
			LogLevel: asynq.FatalLevel,
		},
	})

	redisSync.scheduler = mgr

	if err != nil {
		return nil
	}

	return redisSync
}

func NewWorkerRedisSyncLayer(cfg SyncPlaneSettings, rCfg resources.ResourceSettings) *RedisSync {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,

		MaxRetries: cfg.Redis.MaxRetries,
		PoolSize:   cfg.Redis.PoolSize,
	})

	rm, err := resources.CreateResourceManager(rCfg)
	if err != nil {
		panic(fmt.Sprintf("unable to create resource manager: %s", err))
	}

	queueMap := rm.GetValidTierMap()

	asynqClient := asynq.NewClientFromRedisClient(rdb)
	asynqServer := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: 10,
		Queues:      queueMap,

		LogLevel: asynq.FatalLevel,
	})

	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	redisSync := &RedisSync{
		redisClient: rdb,
		asynqClient: asynqClient,
		asynqServer: asynqServer,
		redsync:     rs,

		mux:                  asynq.NewServeMux(),
		workerResourceManger: rm,

		log: logging.ForService("worker-synclayer"),
	}

	return redisSync
}

func (r *RedisSync) RegisterExecuteNodeHandler(handler ExecuteNodeHandler) error {
	if r.mux == nil {
		return fmt.Errorf("cannot register handler: mux not initalized")
	}

	r.mux.HandleFunc(TypeNodeExecute, func(ctx context.Context, t *asynq.Task) error {
		var p NodeExecutePayload
		err := json.Unmarshal(t.Payload(), &p)
		if err != nil {
			return fmt.Errorf("ExecuteNodeHandler: Error unmarshaling payload: %w", err)
		}

		attempt, _ := asynq.GetRetryCount(ctx)
		p.Attempt = attempt + 1

		if err := handler(ctx, p); err != nil {
			return err
		}

		return nil
	})

	return nil
}

func (r *RedisSync) EnqueueExecuteNode(payload NodeExecutePayload) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	queue := payload.NodeDef.Tier

	r.log.Debug("enqueued node def", "tier", payload.NodeDef.Tier)

	task := asynq.NewTask(TypeNodeExecute, p, asynq.Queue(queue), asynq.MaxRetry(payload.MaxAttempts-1))
	if _, err := r.asynqClient.Enqueue(task); err != nil {
		return err
	}

	return nil
}

func (r *RedisSync) RegisterNodeFinishedHandler(handler NodeFinishedHandler) error {
	if r.mux == nil {
		return fmt.Errorf("cannot register handler: mux not initalized")
	}

	r.mux.HandleFunc(TypeNodeFinished, func(ctx context.Context, t *asynq.Task) error {
		var p NodeFinishedPayload
		err := json.Unmarshal(t.Payload(), &p)
		if err != nil {
			return fmt.Errorf("RegisterExecuteNodeHandler: Error unmarshaling payload: %w", err)
		}
		return handler(ctx, p)
	})

	return nil
}

func (r *RedisSync) EnqueueNodeFinished(payload NodeFinishedPayload) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeNodeFinished, p, asynq.Queue("controller"))
	if _, err := r.asynqClient.Enqueue(task); err != nil {
		return err
	}

	return nil
}

func (r *RedisSync) RegisterNodeFailedHandler(handler NodeFailedHandler) error {
	if r.mux == nil {
		return fmt.Errorf("cannot register handler: mux not initalized")
	}

	r.mux.HandleFunc(TypeNodeFailed, func(ctx context.Context, t *asynq.Task) error {
		var p NodeFailedPayload
		err := json.Unmarshal(t.Payload(), &p)
		if err != nil {
			return fmt.Errorf("RegisterNodeFailedHandler: Error unmarshaling payload: %w", err)
		}
		return handler(ctx, p)
	})

	return nil
}

func (r *RedisSync) EnqueueNodeFailed(payload NodeFailedPayload) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeNodeFailed, p, asynq.Queue("controller"))
	if _, err := r.asynqClient.Enqueue(task); err != nil {
		return err
	}

	return nil
}

func (r *RedisSync) UpdateSubscribedQueues(queues map[string]int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.asynqServer.SetQueues(queues, false)
	return nil

}

const schedulerRunKey = "pup:sched:active_runs"

func (r *RedisSync) RegisterFlowStepHandler(handler FlowStepHandler) error {
	if r.mux == nil {
		return fmt.Errorf("cannot register handler: mux not initalized")
	}

	r.mux.HandleFunc(TypeFlowStep, func(ctx context.Context, t *asynq.Task) error {
		var p FlowStepPayload
		err := json.Unmarshal(t.Payload(), &p)
		if err != nil {
			return fmt.Errorf("RegisterExecuteNodeHandler: Error unmarshaling payload: %w", err)
		}
		return handler(ctx, p)
	})

	return nil
}

const SchedulerElectionInterval = 30 * time.Second

func (r *RedisSync) StartScheduler(ctx context.Context) {

	mutex := r.redsync.NewMutex("asynq:periodic:leader", redsync.WithExpiry(SchedulerElectionInterval), redsync.WithTries(1))

	ticker := time.NewTicker(SchedulerElectionInterval / 2)
	defer ticker.Stop()

	leader := false

	tryBecomeLeader := func(ctx context.Context) {
		if leader {
			return
		}

		if err := mutex.LockContext(ctx); err != nil {
			// couldn't get the lock now; will retry on next tick
			return
		}

		if err := r.scheduler.Start(); err != nil {
			// failed to start manager; release lock and don't mark leader
			mutex.UnlockContext(ctx)
			return
		}

		leader = true
	}

	tryBecomeLeader(ctx)

	for {
		select {
		case <-ctx.Done():
			r.scheduler.Shutdown()
			mutex.UnlockContext(ctx)
			return

		case <-ticker.C:
		}

		if !leader {
			tryBecomeLeader(ctx)
			continue

		}

		ok, err := mutex.ExtendContext(ctx)
		if err != nil || !ok {
			r.scheduler.Shutdown()
			mutex.UnlockContext(ctx)
			leader = false
		}

	}
}
func (r *RedisSync) StopScheduler(ctx context.Context) {
	r.scheduler.Shutdown()
}

func (r *RedisSync) AddRunToScheduler(run_id string) error {
	return r.redisClient.SAdd(context.TODO(), schedulerRunKey, run_id).Err()
}

func (r *RedisSync) RemoveRunFromScheduler(run_id string) error {
	return r.redisClient.SRem(context.TODO(), schedulerRunKey, run_id).Err()
}

type RedisMutex struct {
	mutex *redsync.Mutex
}

func (r *RedisSync) NewMutex(run_id string, duration time.Duration) Mutex {
	key := fmt.Sprintf("redismutex:%s", run_id)

	return &RedisMutex{
		mutex: r.redsync.NewMutex(key, redsync.WithExpiry(duration)),
	}
}

func (mutex *RedisMutex) Lock(ctx context.Context) error {
	return mutex.mutex.LockContext(ctx)
}

func (mutex *RedisMutex) Unlock(ctx context.Context) error {
	_, err := mutex.mutex.UnlockContext(ctx)
	return err
}

func (r *RedisSync) Start() error {

	go func() {
		r.asynqServer.Start(r.mux)
	}()

	go func() {
		if r.scheduler != nil {
			r.StartScheduler(context.Background())
		}
	}()

	return nil
}

func (r *RedisSync) Close() error {

	r.asynqServer.Shutdown()
	if r.scheduler != nil {
		r.scheduler.Shutdown()
	}

	return nil
}

func (r *RedisSync) listScheduledTasks() ([]string, error) {
	return r.redisClient.SMembers(context.TODO(), schedulerRunKey).Result()
}

type schedulerTaskProvider struct {
	sync     *RedisSync
	cronspec string
}

func newSchedulerTaskProvider(sync *RedisSync, cronspec string) *schedulerTaskProvider {
	return &schedulerTaskProvider{
		sync:     sync,
		cronspec: cronspec,
	}
}

func (p *schedulerTaskProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var configs []*asynq.PeriodicTaskConfig

	ids, err := p.sync.listScheduledTasks()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {

		task, err := newFlowStepTask(id)
		if err != nil {
			continue
		}

		configs = append(configs, &asynq.PeriodicTaskConfig{Task: task, Cronspec: p.cronspec})
	}

	return configs, nil
}

func newFlowStepTask(runID string) (*asynq.Task, error) {
	payload, err := json.Marshal(FlowStepPayload{
		RunID: runID,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeFlowStep, payload, asynq.TaskID(runID), asynq.Queue("controller")), nil
}
