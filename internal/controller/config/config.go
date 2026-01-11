package config

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"

	"github.com/pelletier/go-toml/v2"
)

type ControllerConfig struct {
	SyncLayer syncplane.SyncPlaneSettings
	Telemetry telemetry.TelemetrySettings

	Storage struct {
		DataPath string
	}

	Redis struct {
		Address string
	}
}

//go:embed defaults/controller.toml
var defaut_controller_toml string

func resolveConfigPath(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}

	if env := os.Getenv("PUPLOAD_CONFIG"); env != "" {
		return env
	}

	return "/etc/pupload/"
}

func LoadControllerConfig(flagVal string) ControllerConfig {

	configPath := resolveConfigPath(flagVal)

	controller_toml_path := filepath.Join(configPath, "controller.toml")

	if _, err := os.Stat(controller_toml_path); err != nil {
		os.MkdirAll(configPath, 0755)
		os.WriteFile(controller_toml_path, []byte(defaut_controller_toml), 0755)
	}

	controller_toml, err := os.ReadFile(controller_toml_path)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	var controller_config ControllerConfig
	if err := toml.Unmarshal(controller_toml, &controller_config); err != nil {
		log.Fatalln("Unable to load config", err)
	}

	return controller_config

}
