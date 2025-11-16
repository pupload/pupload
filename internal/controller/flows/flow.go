package flows

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

type FlowService struct {
	FlowPath string
	NodeDefs map[string]NodeDef
	FlowList map[string]Flow
}

type Flow struct {
	AvailableNodes []string
	Nodes          []Node
}

//go:embed defaults/Flows/*.yaml
var defaultFlows embed.FS

//go:embed defaults/NodeDefs/*.yaml
var defaultNodeDefs embed.FS

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

func CreateFlowService(dataPath string) FlowService {
	flowPath := filepath.Join(dataPath, "Flows")
	nodeDefPath := filepath.Join(dataPath, "NodeDefs")

	if _, err := os.Stat(flowPath); err != nil {
		os.MkdirAll(flowPath, 0755)
		writeDefaultFlows(flowPath)
	}

	if _, err := os.Stat(nodeDefPath); err != nil {
		os.MkdirAll(nodeDefPath, 0755)
	}

	flowYamls, err := os.ReadDir(flowPath)
	if err != nil {
		log.Fatalln("Can't read flow path", err)
	}

	flowMap := make(map[string]Flow)

	for _, e := range flowYamls {
		var flow Flow
		data, err := os.ReadFile(filepath.Join(flowPath, e.Name()))
		if err != nil {
			log.Println("Could not read file: ", e.Name(), err)
			continue
		}
		if err := yaml.Unmarshal(data, &flow); err != nil {
			log.Println("Could not unmarshal yaml: ", e.Name(), err)
			continue
		}

		log.Println(flow)
		flowMap[e.Name()] = flow
	}

	nodeDefYamls, err := os.ReadDir(nodeDefPath)
	if err != nil {
		log.Fatalln("Can't read node definition path", err)
	}

	nodeDefMap := make(map[string]NodeDef)

	for _, e := range nodeDefYamls {
		var nodeDef NodeDef
		data, err := os.ReadFile(filepath.Join(nodeDefPath))

		if err != nil {
			log.Println("Could not read file: ", e.Name(), err)
			continue
		}

		if err := yaml.Unmarshal(data, &nodeDef); err != nil {
			log.Println("Could not unmarshal yaml: ", e.Name(), err)
			continue
		}

		nodeDefMap[e.Name()] = nodeDef
	}

	return FlowService{
		FlowPath: flowPath,
		FlowList: flowMap,
	}

}
