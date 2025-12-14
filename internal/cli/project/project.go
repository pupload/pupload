package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"pupload/internal/models"

	"sigs.k8s.io/yaml"
)

type ControllerDef struct {
	Name string
	URL  string
}

type ProjectFile struct {
	ProjectName string `yaml:"ProjectName"`
	Version     int    `yaml:"Version"`

	Controllers  []ControllerDef     `yaml:"Controllers"`
	GlobalStores []models.StoreInput `yaml:"GlobalStores"`

	Extra map[string]any `yaml:",inline"`
}

func defaultProjectFile(projectName string) ProjectFile {
	return ProjectFile{
		ProjectName:  projectName,
		Version:      0,
		Controllers:  []ControllerDef{},
		GlobalStores: []models.StoreInput{},
	}
}

func TestFlow(projectRoot, controllerAddress, flowName string) (*models.FlowRun, error) {

	flow, err := GetFlow(projectRoot, flowName)
	if err != nil {
		return nil, err
	}

	node_defs, err := GetNodeDefs(projectRoot)
	if err != nil {
		return nil, err
	}

	body := struct {
		Flow     models.Flow      `json:"Flow"`
		NodeDefs []models.NodeDef `json:"NodeDefs"`
	}{
		Flow:     *flow,
		NodeDefs: node_defs,
	}

	j, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	url, _ := url.JoinPath(controllerAddress, "api", "v1", "flow", "test")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	flow_run := new(models.FlowRun)
	err = json.NewDecoder(resp.Body).Decode(flow_run)

	return flow_run, nil
}

func GetFlow(projectRoot, flowName string) (*models.Flow, error) {

	flows, _ := GetFlows(projectRoot)

	var flow *models.Flow

	for _, f := range flows {
		if f.Name == flowName {
			flow = &f
			break
		}
	}

	if flow == nil {
		return nil, fmt.Errorf("flow %s not found", flowName)
	}

	return flow, nil
}

func GetNodeDefs(projectRoot string) ([]models.NodeDef, error) {
	path := filepath.Join(projectRoot, "node_defs")
	nodeDefs := make([]models.NodeDef, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var nodeDef models.NodeDef
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &nodeDef); err != nil {
			continue
		}

		nodeDefs = append(nodeDefs, nodeDef)
	}

	return nodeDefs, nil
}

func GetFlows(projectRoot string) ([]models.Flow, error) {
	path := filepath.Join(projectRoot, "flows")
	flows := make([]models.Flow, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var flow models.Flow
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &flow); err != nil {
			continue
		}

		flows = append(flows, flow)
	}

	return flows, nil
}

func GetProjectFile(path string) (*ProjectFile, error) {
	rootPath, err := ProjectRoot(path)
	if err != nil {
		return nil, err
	}

	projectPath := filepath.Join(rootPath, "pup.yaml")
	file, err := os.ReadFile(projectPath)
	if err != nil {
		return nil, err
	}

	var project ProjectFile
	if err := yaml.Unmarshal(file, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

func InitProject(path string, projectName string) error {

	file := filepath.Join(path, "pup.yaml")

	project := defaultProjectFile(projectName)
	yamlBytes, yErr := yaml.Marshal(&project)
	if yErr != nil {
		return yErr
	}

	err := os.WriteFile(file, yamlBytes, 0755)
	if err != nil {
		return err
	}

	if err := os.Mkdir("flows", 0755); err != nil {
		return err
	}

	if err := os.Mkdir("node_defs", 0755); err != nil {
		return err
	}

	return nil

}

func IsProjectDirectory(path string) bool {
	if _, err := ProjectRoot(path); err != nil {
		return false
	}

	return true
}

func GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := ProjectRoot(dir)
	if err != nil {
		return "", err
	}

	return root, nil

}

func ProjectRoot(path string) (string, error) {

	for {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}

		for _, e := range entries {
			if e.Name() == "pup.yaml" {
				return filepath.Abs(path)
			}
		}

		path = filepath.Join(path, "..")
	}
}
