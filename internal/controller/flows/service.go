package flows

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"pupload/internal/logging"
	"pupload/internal/models"
	"pupload/internal/stores"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"sigs.k8s.io/yaml"
)

//go:embed defaults/Flows/*.yaml
var defaultFlows embed.FS

//go:embed defaults/NodeDefs/*.yaml
var defaultNodeDefs embed.FS

type LocalStoreKey [2]string

type FlowService struct {
	FlowPath       string
	NodeDefs       map[string]models.NodeDef
	FlowList       map[string]models.Flow
	GlobalStoreMap map[string]models.Store        // Maps StoreName onto a given store.
	LocalStoreMap  map[LocalStoreKey]models.Store // Maps [FlowName, StoreName] onto a given store

	RedisClient *redis.Client
	AsynqClient *asynq.Client
	AsynqServer *asynq.Server

	log *slog.Logger
}

func CreateFlowService(dataPath string, rdb *redis.Client) *FlowService {

	localStoreMap := make(map[LocalStoreKey]models.Store)
	globalStoreMap := make(map[string]models.Store)
	flowMap := make(map[string]models.Flow)
	nodeDefMap := make(map[string]models.NodeDef)

	flowPath := filepath.Join(dataPath, "Flows")
	nodeDefPath := filepath.Join(dataPath, "NodeDefs")

	asynqClient := asynq.NewClientFromRedisClient(rdb)
	asynqServer := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"controller": 1,
		},
	})

	f := FlowService{
		FlowPath:       flowPath,
		FlowList:       flowMap,
		NodeDefs:       nodeDefMap,
		RedisClient:    rdb,
		AsynqClient:    asynqClient,
		AsynqServer:    asynqServer,
		LocalStoreMap:  localStoreMap,
		GlobalStoreMap: globalStoreMap,
		log:            logging.ForService("flow"),
	}

	if _, err := os.Stat(flowPath); err != nil {
		os.MkdirAll(flowPath, 0755)
		writeDefaultFlows(flowPath)
	}

	if _, err := os.Stat(nodeDefPath); err != nil {
		os.MkdirAll(nodeDefPath, 0755)
		writeDefaultNodeDefs(nodeDefPath)
	}

	flowYamls, err := os.ReadDir(flowPath)
	if err != nil {
		log.Fatalln("Can't read flow path", err)
	}

	for _, e := range flowYamls {
		data, err := os.ReadFile(filepath.Join(flowPath, e.Name()))
		if err != nil {
			log.Println("Could not read file: ", e.Name(), err)
			continue
		}

		flow, err := f.processAndValidateFlow(data, e.Name())

		if err != nil {
			log.Println("Could not process flow: ", e.Name(), err)
			continue
		}

		flowMap[e.Name()] = flow
	}

	nodeDefYamls, err := os.ReadDir(nodeDefPath)
	if err != nil {
		log.Fatalln("Can't read node definition path", err)
	}

	for _, e := range nodeDefYamls {
		var nodeDef models.NodeDef
		data, err := os.ReadFile(filepath.Join(nodeDefPath, e.Name()))

		if err != nil {
			log.Println("Could not read file: ", e.Name(), err)
			continue
		}

		if err := yaml.Unmarshal(data, &nodeDef); err != nil {
			log.Println("Could not unmarshal yaml: ", e.Name(), err)
			continue
		}

		defName := fmt.Sprintf("%s/%s", nodeDef.Publisher, nodeDef.Name)
		nodeDefMap[defName] = nodeDef

	}

	go func() {
		asynqServer.Start(f.AsynqConfigureHandlers())
	}()

	f.log.Info("Flows", "flows", f.FlowList)

	return &f
}

func (f *FlowService) Close() {
	f.AsynqServer.Stop()
	f.AsynqClient.Close()
	f.RedisClient.Close()
}

func (f *FlowService) ValidateFlow(flow models.Flow) error {
	// Check that all node definitions exist

	// Check that all edges are valid

	// Check for cycles

	// Check that all stores are valid

	// Check that all data wells are valid
	for _, well := range flow.DataWells {
		if !(well.Type == "static" || well.Type == "dynamic") {
			return fmt.Errorf("Data well %s has invalid type %s", well.Edge, well.Type)
		}

		if well.Type == "static" && well.Key == nil {
			return fmt.Errorf("Data well %s is static but has no key", well.Edge)
		}

		if well.Type == "dynamic" && well.Key != nil && validateKey(*well.Key) {
			return fmt.Errorf("Data well %s key is invalid", well.Edge)
		}

	}

	return nil
}

func validateKey(key string) bool {
	return len(key) > 0
}

func (f *FlowService) processAndValidateFlow(data []byte, name string) (models.Flow, error) {
	var flow models.Flow
	if err := yaml.Unmarshal(data, &flow); err != nil {
		return flow, fmt.Errorf("Could not process flow: %s", err)
	}

	for _, storeInput := range flow.Stores {
		store, err := stores.UnmarshalStore(storeInput)
		if err != nil {
			f.log.Warn("Invalid store definition", "flow", name, "store", storeInput.Name, "error", err.Error())
			continue
		}

		f.LocalStoreMap[LocalStoreKey{name, storeInput.Name}] = store
	}

	validateErr := f.ValidateFlow(flow)
	if validateErr != nil {
		return flow, validateErr
	}

	return flow, nil

}

func writeDefaultFlows(flowPath string) {
	fs.WalkDir(defaultFlows, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		yamlFile, err := defaultFlows.ReadFile(path)
		if err != nil {
			return err
		}
		writePath := filepath.Join(flowPath, d.Name())
		os.WriteFile(writePath, yamlFile, 0755)

		return nil

	})
}

func writeDefaultNodeDefs(nodeDefPath string) {
	fs.WalkDir(defaultNodeDefs, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		yamlFile, err := defaultNodeDefs.ReadFile(path)
		if err != nil {
			return err
		}
		writePath := filepath.Join(nodeDefPath, d.Name())
		os.WriteFile(writePath, yamlFile, 0755)

		return nil

	})
}
