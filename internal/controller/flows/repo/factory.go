package repo

import (
	"fmt"

	"github.com/pupload/pupload/internal/controller/flows/repo/project"
	runtime_repo "github.com/pupload/pupload/internal/controller/flows/repo/runtime"

	"github.com/redis/go-redis/v9"
)

type ProjectRepoType string

const (
	NoneProject     ProjectRepoType = "none"
	SingleProjectFS ProjectRepoType = "SingleProjectFS"
)

type ProjectRepoSettings struct {
	Type ProjectRepoType

	SingleProjectFS SingleProjectFSSettings
}

type SingleProjectFSSettings struct {
	WorkingDir string
}

func CreateProjectRepo(cfg ProjectRepoSettings) (ProjectRepo, error) {
	switch cfg.Type {
	case SingleProjectFS:
		// if !fs.ValidPath(cfg.SingleProjectFS.WorkingDir) {
		// 	return nil, fmt.Errorf("specified working directory is invalid")
		// }

		return project.CreateSingleProjectFs(cfg.SingleProjectFS.WorkingDir), nil
	}

	return nil, fmt.Errorf("invalid project repo config")
}

type RuntimeRepoType string

const (
	RedisRuntimeRepo RuntimeRepoType = "redis"
)

type RuntimeRepoSettings struct {
	Type RuntimeRepoType

	Redis RedisSettings
}

type RedisSettings struct {
	Address  string
	Password string
	DB       int
}

func CreateRuntimeRepo(cfg RuntimeRepoSettings) (RuntimeRepo, error) {
	switch cfg.Type {
	case RedisRuntimeRepo:
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		return runtime_repo.CreateRedisRuntimeRepo(rdb), nil
	}

	return nil, fmt.Errorf("invalid runtime repo config")
}
