package server

import (
	"fmt"

	"github.com/pupload/pupload/internal/resources"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/worker/container"
	"github.com/pupload/pupload/internal/worker/node"
)

func NewWorkerServer(s syncplane.SyncLayer, cs *container.ContainerService, rm *resources.ResourceManager) {

	ns, err := node.CreateNodeService(cs, s, rm)
	if err != nil {
		panic(fmt.Sprintf("Unable to create node service: %s", err))
	}

	s.RegisterExecuteNodeHandler(ns.FinishedMiddleware)
	s.Start()
}
