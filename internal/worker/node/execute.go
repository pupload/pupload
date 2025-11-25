package node

import (
	"context"
	"encoding/json"
	"fmt"
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

	ups := n.GetInputStreams(p.InputURLs, "/tmp")
	for val, key := range ups {
		fmt.Println(ups)
		env = append(env, fmt.Sprintf("%s=%s", val, key.path))
	}

	res, err := n.ContainerService.DockerClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Env: env,
			Cmd: []string{"-h"},
		},
		HostConfig: &container.HostConfig{
			AutoRemove: false,
		},
		Image: p.NodeDef.Image,
		Name:  "test2",
	})

	container_id := res.ID

	if err != nil {
		log.Println(err)
		return err
	}

	for key, val := range ups {
		fmt.Println(key)
		_, err := n.ContainerService.DockerClient.CopyToContainer(ctx, container_id, client.CopyToContainerOptions{
			DestinationPath: val.base_path,
			Content:         val.reader,
		})

		if err != nil {
			fmt.Printf("Error copying url to file: %s", err)
		}

		val.reader.Close()
	}

	if _, err := n.ContainerService.DockerClient.ContainerStart(ctx, container_id, client.ContainerStartOptions{}); err != nil {
		log.Printf("Error Starting docker container: %s", err)
		return err
	}

	n.ContainerService.DockerClient.ContainerWait(context.TODO(), container_id, client.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning,
	})

	return nil
}
