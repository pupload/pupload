package container

import (
	"context"

	"github.com/moby/moby/client"
)

type IImageManager interface {
	Pull()
	Validate()
}

type ImageManager struct {
	client *client.Client
}

func (im *ImageManager) Pull(ctx context.Context, refStr string) error {
	res, err := im.client.ImagePull(context.Background(), refStr, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer res.Close()

	if err := res.Wait(ctx); err != nil {
		return err
	}

	return nil
}

func (im *ImageManager) Validate(ctx context.Context, refStr string) (bool, error) {
	filter := client.Filters{}.Add("reference", refStr)

	res, err := im.client.ImageList(ctx, client.ImageListOptions{
		Filters: filter,
	})

	if err != nil {
		return false, err
	}

	return len(res.Items) > 0, nil
}
