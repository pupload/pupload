package flows

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"sigs.k8s.io/yaml"
)

//go:embed defaults/Flows/*.yaml
var defaultFlows embed.FS

//go:embed defaults/NodeDefs/*.yaml
var defaultNodeDefs embed.FS

type FlowService struct {
	FlowPath    string
	NodeDefs    map[string]models.NodeDef
	FlowList    map[string]models.Flow
	RedisClient *redis.Client
	AsynqClient *asynq.Client
}

func CreateFlowService(dataPath string, redisClient *redis.Client) FlowService {
	flowPath := filepath.Join(dataPath, "Flows")
	nodeDefPath := filepath.Join(dataPath, "NodeDefs")

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

	flowMap := make(map[string]models.Flow)

	for _, e := range flowYamls {
		var flow models.Flow
		data, err := os.ReadFile(filepath.Join(flowPath, e.Name()))
		if err != nil {
			log.Println("Could not read file: ", e.Name(), err)
			continue
		}
		if err := yaml.Unmarshal(data, &flow); err != nil {
			log.Println("Could not unmarshal yaml: ", e.Name(), err)
			continue
		}

		flowMap[e.Name()] = flow
	}

	nodeDefYamls, err := os.ReadDir(nodeDefPath)
	if err != nil {
		log.Fatalln("Can't read node definition path", err)
	}

	nodeDefMap := make(map[string]models.NodeDef)

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

	asynqClient := asynq.NewClientFromRedisClient(redisClient)

	return FlowService{
		FlowPath:    flowPath,
		FlowList:    flowMap,
		NodeDefs:    nodeDefMap,
		RedisClient: redisClient,
		AsynqClient: asynqClient,
	}

}

func (f *FlowService) ListFlows() map[string]models.Flow {
	return f.FlowList
}

func (f *FlowService) StartFlow(name string) (string, error) {
	flow, exists := f.FlowList[name]
	if !exists {
		return "", fmt.Errorf("Flow %s does not exist", name)
	}

	if len(flow.Nodes) == 0 {
		return "", fmt.Errorf("Flow %s does not contain any nodes", name)
	}

	entryNode := flow.Nodes[0]

	id, err := f.createFlowRun(name)

	if err != nil {
		return "", err
	}

	err = f.ExecuteNode(&entryNode)

	if err != nil {
		return "", err
	}

	return id, nil
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
