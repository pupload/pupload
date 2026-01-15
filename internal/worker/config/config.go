package config

import (
	"github.com/pupload/pupload/internal/resources"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

type WorkerConfig struct {
	Worker    WorkerSettings
	SyncPlane syncplane.SyncPlaneSettings
	Telemetry telemetry.TelemetrySettings
	Resources resources.ResourceSettings
	Runtime   RuntimeSettings

	Logging  LoggingSettings
	Security SecuritySettings
}

type WorkerSettings struct {
	ID string
}

type RuntimeSettings struct {
	ContainerEngine  string `json:"container_engine"`  // docker, podman, runsc, auto
	ContainerRuntime string `json:"container_runtime"` // runc, gvisor, auto

	EnableGPUSupport bool `json:"enable_gpu_support"`

	Gvisor GvisorSettings `json:"gvisor"`
}

type GvisorSettings struct {
	Platform string `json:"platform"` // systrap, kvm, ptrace
}

type LoggingSettings struct {
	LogLevel string `json:"log_level"` // debug, info, warn, error

}

type SecuritySettings struct {
	AllowedRegistries []string `json:"allowed_registries"` // all if empty
	ImageAllowlist    []string `json:"image_allowlist"`    //
	ImageBlocklist    []string `json:"image_blocklist"`
}

func DefaultConfig() *WorkerConfig {
	return &WorkerConfig{
		Worker: WorkerSettings{
			ID: "0",
		},

		SyncPlane: syncplane.SyncPlaneSettings{
			SelectedSyncPlane: "redis",
			Redis: syncplane.RedisSettings{
				Address:  "localhost:6379",
				Password: "",
				DB:       0,

				PoolSize:   10,
				MaxRetries: 3,
			},
		},

		Resources: resources.ResourceSettings{
			MaxCPU:     "auto",
			MaxMemory:  "8gb",
			MaxStorage: "50gb",
		},

		Runtime: RuntimeSettings{
			ContainerEngine:  "auto",
			ContainerRuntime: "auto",
			EnableGPUSupport: false,
		},

		Logging: LoggingSettings{
			LogLevel: "warn",
		},

		Telemetry: telemetry.TelemetrySettings{
			Enabled: false,
		},

		Security: SecuritySettings{
			AllowedRegistries: []string{
				"docker.io",
				"ghcr.io",
				"gcr.io",
				"public.ecr.aws",
			},

			ImageAllowlist: []string{},
			ImageBlocklist: []string{},
		},
	}
}
