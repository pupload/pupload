package run

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"pupload/internal/controller"
	controllerconfig "pupload/internal/controller/config"
	"pupload/internal/controller/flows/repo"
	"pupload/internal/syncplane"
	"pupload/internal/worker"
	workerconfig "pupload/internal/worker/config"
	"syscall"

	"github.com/alicebob/miniredis/v2"
	"golang.org/x/sync/errgroup"
)

func RunDev(projectRoot string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s, err := miniredis.Run()
	if err != nil {
		return err
	}

	syncplane := syncplane.SyncPlaneSettings{
		SelectedSyncPlane: "redis",
		Redis: syncplane.RedisSettings{
			Address:  s.Addr(),
			Password: "",

			PoolSize:   10,
			MaxRetries: 3,
		},

		ControllerStepInterval: "@every 10s",
	}

	controller_cfg := controllerconfig.DefaultConfig()
	controller_cfg.SyncPlane = syncplane
	controller_cfg.ProjectRepo = repo.ProjectRepoSettings{
		Type: repo.SingleProjectFS,

		SingleProjectFS: repo.SingleProjectFSSettings{
			WorkingDir: projectRoot,
		},
	}
	controller_cfg.RuntimeRepo = repo.RuntimeRepoSettings{
		Type: repo.RedisRuntimeRepo,

		Redis: repo.RedisSettings{
			Address:  s.Addr(),
			Password: "",
			DB:       0,
		},
	}

	worker_cfg := workerconfig.DefaultConfig()
	worker_cfg.SyncPlane = syncplane

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return controller.RunWithConfig(gctx, controller_cfg)
	})

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return worker.RunWithConfig(gctx, worker_cfg)
	})

	return g.Wait()
}

func RunDevSilent(projectRoot string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s, err := miniredis.Run()
	if err != nil {
		return err
	}

	syncplane := syncplane.SyncPlaneSettings{
		SelectedSyncPlane: "redis",
		Redis: syncplane.RedisSettings{
			Address:  s.Addr(),
			Password: "",

			PoolSize:   10,
			MaxRetries: 3,
		},

		ControllerStepInterval: "@every 10s",
	}

	controller_cfg := controllerconfig.DefaultConfig()
	controller_cfg.SyncPlane = syncplane
	controller_cfg.ProjectRepo = repo.ProjectRepoSettings{
		Type: repo.SingleProjectFS,

		SingleProjectFS: repo.SingleProjectFSSettings{
			WorkingDir: projectRoot,
		},
	}

	controller_cfg.RuntimeRepo = repo.RuntimeRepoSettings{
		Type: repo.RedisRuntimeRepo,

		Redis: repo.RedisSettings{
			Address:  s.Addr(),
			Password: "",
			DB:       0,
		},
	}

	worker_cfg := workerconfig.DefaultConfig()
	worker_cfg.SyncPlane = syncplane

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return controller.RunWithConfigSilent(gctx, controller_cfg)
	})

	g.Go(func() error {
		log.SetOutput(io.Discard)
		return worker.RunWithConfigSilent(gctx, worker_cfg)
	})

	return g.Wait()
}
