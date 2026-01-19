package container

import (
	"context"
	"log"

	"github.com/moby/moby/client"
)

type ContainerService struct {
	DockerClient *client.Client
	RT           *ContainerRuntime
	IO           *ContainerIO
	IM           *ImageManager
}

func CreateContainerService() ContainerService {
	client, err := client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Can't create container service\n%s", err.Error())
	}

	return ContainerService{
		DockerClient: client,
		RT: &ContainerRuntime{
			client:  client,
			runtime: "runc",
		},

		IO: &ContainerIO{
			client: client,
		},

		IM: &ImageManager{
			client: client,
		},
	}

}

func (cs *ContainerService) ListImages() ([]string, error) {

	res, err := cs.DockerClient.ImageList(context.TODO(), client.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	imageNames := make([]string, 0)
	for _, img := range res.Items {
		imageNames = append(imageNames, img.RepoTags...)
	}

	return imageNames, nil

}
