package syncplane

import (
	"context"
	"pupload/internal/models"

	"github.com/moby/moby/api/types/container"
)

const (
	TypeFlowStep        = "flow:step"
	TypeNodeExecute     = "node:execute"
	TypeNodeFinished    = "node:finished"
	TypeNodeError       = "node:error"
	TypeControllerClean = "controller:clean"
)

type ExecuteNodeHandler func(ctx context.Context, payload NodeExecutePayload, resource container.Resources) error
type NodeExecutePayload struct {
	RunID      string
	NodeDef    models.NodeDef
	Node       models.Node
	InputURLs  map[string]string
	OutputURLs map[string]string
}

type FlowStepHandler func(ctx context.Context, payload FlowStepPayload) error
type FlowStepPayload struct {
	RunID string
}

type NodeFinishedHandler func(ctx context.Context, payload NodeFinishedPayload) error
type NodeFinishedPayload struct {
	RunID  string
	NodeID string
	Logs   []models.LogRecord
}

type NodeErrorHandler func(ctx context.Context, payload NodeErrorPayload) error
type NodeErrorPayload struct {
}
