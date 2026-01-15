package server

import (
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/worker/container"
	"github.com/pupload/pupload/internal/worker/node"
)

func NewWorkerServer(s syncplane.SyncLayer, cs *container.ContainerService) {

	ns := node.CreateNodeService(cs, s)

	s.RegisterExecuteNodeHandler(ns.FinishedMiddleware)
	s.Start()
}
