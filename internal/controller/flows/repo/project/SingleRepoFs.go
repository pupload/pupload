package project

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pupload/pupload/internal/models"

	"go.yaml.in/yaml/v2"
)

type SingleProjectFs struct {
	wd string
}

func CreateSingleProjectFs(workingDir string) *SingleProjectFs {
	return &SingleProjectFs{
		wd: workingDir,
	}
}

func (r *SingleProjectFs) SaveProject(ctx context.Context, project models.Project) error {
	return fmt.Errorf("can't save single project repo, this is CLI's or other tools responsibility")
}

func (r *SingleProjectFs) LoadProject(ctx context.Context, tenantID, projectName string) (models.Project, error) {

	flows, err := loadFlowsFromDir(path.Join(r.wd, "flows"))
	if err != nil {
		return models.Project{}, err
	}

	defs, err := loadDefsFromDir(path.Join(r.wd, "node_defs"))
	if err != nil {
		return models.Project{}, err
	}

	return models.Project{
		TenantID:    "global",
		ProjectName: "single",
		Version:     1,

		Flows:        flows,
		NodeDefs:     defs,
		GlobalStores: []models.StoreInput{},
	}, nil
}

func (r *SingleProjectFs) DeleteProject(ctx context.Context, tenantID, projectName string) error {
	return fmt.Errorf("single project repos projects can not be deleted")
}

func (r *SingleProjectFs) Close(ctx context.Context) error {
	return fmt.Errorf("single project repos projects can not be deleted")
}

func loadFlowsFromDir(path string) ([]models.Flow, error) {
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

func loadDefsFromDir(path string) ([]models.NodeDef, error) {
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
