package node

import (
	"fmt"
	"pupload/internal/models"
	"pupload/internal/worker/container"
	"slices"
)

type NodeService struct {
	ContainerService *container.ContainerService
}

func CreateNodeService(cs *container.ContainerService) NodeService {
	return NodeService{
		ContainerService: cs,
	}
}

func (ns NodeService) GetEnvArray(nodeDef models.NodeDef, node models.Node) ([]string, error) {

	env := make([]string, 0)

	flags, err := ns.GetFlags(nodeDef, node)
	if err != nil {
		return []string{}, err
	}

	// TODO: ensure this can't be escaped
	for key, val := range flags {
		env = append(env, key+"="+val)
	}

	return env, nil
}

func (ns NodeService) GetFlags(nodeDef models.NodeDef, node models.Node) (map[string]string, error) {
	flagMap := make(map[string]string)

	for _, flagDef := range nodeDef.Flags {

		for _, flagVal := range node.Flags {
			if flagVal.Name == flagDef.Name {
				flagMap[flagVal.Name] = flagVal.Value
				break
			}
		}

		if _, ok := flagMap[flagDef.Name]; !ok && flagDef.Required {
			return flagMap, fmt.Errorf("Flag %s is required", flagDef.Name)
		}

	}

	return flagMap, nil
}

func (ns NodeService) CanWorkerRunContainer(nodeDef models.NodeDef) bool {

	imageList, err := ns.ContainerService.ListImages()
	if err != nil {
		return false
	}

	return slices.Contains(imageList, nodeDef.Image)

}
