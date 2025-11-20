package node

import (
	"context"
	"encoding/json"
	"log"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (n NodeService) HandleNodeExecuteTask(ctx context.Context, t *asynq.Task) error {
	var p models.NodeExecutePayload

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	capable := n.CanWorkerRunContainer(p.NodeDef)
	if !capable {

	}

	env, err := n.GetEnvArray(p.NodeDef, p.Node)
	if err != nil {
		log.Println(err)
		return err
	}

	res, err := n.ContainerService.DockerClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Env: env,
			Cmd: []string{"-h"},
		},
		Image: p.NodeDef.Image,
		Name:  "test",
	})

	container_id := res.ID
	defer n.ContainerService.DockerClient.ContainerRemove(ctx, container_id, client.ContainerRemoveOptions{
	 	RemoveVolumes: true,
	})

	if err != nil {
		log.Println(err)
		return err
	}

	n.ContainerService.DockerClient.CopyToContainer(ctx, container_id, client.CopyToContainerOptions{})

	if _, err := n.ContainerService.DockerClient.ContainerStart(ctx, container_id, client.ContainerStartOptions{}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
