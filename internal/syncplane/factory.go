package syncplane

import (
	"fmt"
	"pupload/internal/resources"
)

func CreateControllerSyncLayer(cfg SyncPlaneSettings) (SyncLayer, error) {
	switch cfg.SelectedSyncPlane {
	case "redis":
		return NewControllerRedisSyncLayer(cfg), nil
	}

	return nil, fmt.Errorf("")
}

func CreateWorkerSyncLayer(cfg SyncPlaneSettings, rCfg resources.ResourceSettings) (SyncLayer, error) {

	switch cfg.SelectedSyncPlane {
	case "redis":
		return NewWorkerRedisSyncLayer(cfg, rCfg), nil
	}
	return nil, fmt.Errorf("")
}
