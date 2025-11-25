package models

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type QueueClients struct {
	RedisClient *redis.Client
	AsynqClient *asynq.Client
	AsynqServer *asynq.Server
}
