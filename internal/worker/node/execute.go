package node

import (
	"context"
	"fmt"
	"os"
	"pupload/internal/logging"
	"pupload/internal/models"
	"pupload/internal/syncplane"
	"strings"

	cont "pupload/internal/worker/container"

	"github.com/moby/moby/api/types/container"
	"golang.org/x/sync/errgroup"
)

func (n *NodeService) NodeExecute(ctx context.Context, payload syncplane.NodeExecutePayload, resource container.Resources) error {

	l := logging.LoggerFromCtx(ctx)

	// handle worker capabiliites

	in, out, err := n.prepareIO(payload.InputURLs, payload.OutputURLs, payload.NodeDef, "/tmp")
	if err != nil {
		return err
	}

	command, err := n.generateCommand(payload.Node, payload.NodeDef, in, out)
	if err != nil {
		return err
	}

	containerID, err := n.CS.RT.CreateContainer(ctx, cont.ContainerConfig{
		Image: payload.NodeDef.Image,
		Name:  "test2",
		Cmd:   command,

		HostConfig: &container.HostConfig{
			AutoRemove: false,
			Resources:  resource,
		},
	})

	if err != nil {
		return err
	}

	l.With("container_id", containerID)
	l.Info("container created")

	defer n.CS.RT.RemoveContainer(ctx, containerID)

	if err := n.downloadAllInputsToContainer(ctx, containerID, in); err != nil {
		return err
	}

	l.Info("files downloaded to container")

	if err := n.CS.RT.StartContainer(ctx, containerID); err != nil {
		return err
	}

	l.Info("container started")

	res, err := n.CS.RT.WaitContainer(ctx, containerID)
	if err != nil {
		return err
	}

	l.With(
		"exit_code", res.ExitCode,
		"exit_message", res.Error,
	)
	l.Info("container finished")

	logs, err := n.CS.RT.GetLogs(ctx, containerID)
	if err != nil {
		return err
	}

	l.Debug("container logs", "logs", logs)

	if res.ExitCode != 0 {
		return fmt.Errorf("contained exited with non-0 exit code")
	}

	if err := n.uploadAllOutputsFromContainer(ctx, containerID, out); err != nil {
		return err
	}

	l.Info("files uploaded from container")

	return nil
}

func (n *NodeService) downloadAllInputsToContainer(ctx context.Context, containerID string, inputs []preparedIO) error {
	inGroup, errCtx := errgroup.WithContext(ctx)
	for _, i := range inputs {
		i := i
		inGroup.Go(func() error {
			return n.CS.IO.DownloadIntoContainer(errCtx, containerID, i.url, i.base_path, i.filename)
		})
	}

	return inGroup.Wait()
}

func (n *NodeService) uploadAllOutputsFromContainer(ctx context.Context, containerID string, outputs []preparedIO) error {
	outGroup, errCtx := errgroup.WithContext(ctx)
	for _, o := range outputs {
		o := o
		outGroup.Go(func() error {
			return n.CS.IO.UploadFromContainer(errCtx, containerID, o.url, o.path, o.filename)
		})
	}

	return outGroup.Wait()
}

func (n *NodeService) generateCommand(node models.Node, nodeDef models.NodeDef, in, out []preparedIO) ([]string, error) {
	envMap := make(map[string]string)

	if err := n.AddEnvFlagMap(envMap, nodeDef, node); err != nil {
		return nil, err
	}

	// prep inputs
	n.AddIOToEnvMap(envMap, in)
	n.AddIOToEnvMap(envMap, out)

	expand := os.Expand(nodeDef.Command.Exec, func(s string) string {
		return envMap[s]
	})

	command := strings.Fields(expand)
	return command, nil
}
