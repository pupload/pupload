package config

import (
	_ "embed"
	"os"

	"github.com/pupload/pupload/internal/controller/flows/repo"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

type ControllerSettings struct {
	SyncPlane syncplane.SyncPlaneSettings
	Telemetry telemetry.TelemetrySettings

	ProjectRepo repo.ProjectRepoSettings
	RuntimeRepo repo.RuntimeRepoSettings

	Storage struct {
		DataPath string
	}
}

func DefaultConfig() *ControllerSettings {

	wd, err := os.Getwd()
	if err != nil {
		wd = ""
	}

	return &ControllerSettings{
		SyncPlane: syncplane.SyncPlaneSettings{
			SelectedSyncPlane: "redis",
			Redis: syncplane.RedisSettings{
				Address:  "localhost:6379",
				Password: "",
				DB:       0,

				PoolSize:   10,
				MaxRetries: 3,
			},

			ControllerStepInterval: "@every 10s",
		},

		ProjectRepo: repo.ProjectRepoSettings{
			Type: repo.SingleProjectFS,

			SingleProjectFS: repo.SingleProjectFSSettings{
				WorkingDir: wd,
			},
		},

		RuntimeRepo: repo.RuntimeRepoSettings{
			Type: repo.RedisRuntimeRepo,
			Redis: repo.RedisSettings{
				Address:  "localhost:6379",
				Password: "",
				DB:       0,
			},
		},

		Telemetry: telemetry.TelemetrySettings{
			Enabled: false,
		},
	}

}

func resolveConfigPath(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}

	if env := os.Getenv("PUPLOAD_CONFIG"); env != "" {
		return env
	}

	return "/etc/pupload/"
}

// func LoadControllerConfig(flagVal string) ControllerConfig {

// 	configPath := resolveConfigPath(flagVal)

// 	controller_toml_path := filepath.Join(configPath, "controller.toml")

// 	if _, err := os.Stat(controller_toml_path); err != nil {
// 		os.MkdirAll(configPath, 0755)
// 		// os.WriteFile(controller_toml_path, []byte(defaut_controller_toml), 0755)
// 	}

// 	controller_toml, err := os.ReadFile(controller_toml_path)
// 	if err != nil {
// 		log.Fatalln("Unable to load config", err)
// 	}

// 	var controller_config ControllerConfig
// 	if err := toml.Unmarshal(controller_toml, &controller_config); err != nil {
// 		log.Fatalln("Unable to load config", err)
// 	}

// 	return controller_config

// }
