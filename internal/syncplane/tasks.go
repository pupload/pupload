package syncplane

import (
	"context"

	"github.com/pupload/pupload/internal/models"
)

const (
	TypeFlowStep        = "flow:step"
	TypeNodeExecute     = "node:execute"
	TypeNodeFinished    = "node:finished"
	TypeNodeFailed      = "node:failed"
	TypeControllerClean = "controller:clean"
)

type ExecuteNodeHandler func(ctx context.Context, payload NodeExecutePayload) error
type NodeExecutePayload struct {
	RunID      string
	NodeDef    models.NodeDef
	Node       models.Node
	InputURLs  map[string]string
	OutputURLs map[string]string

	MaxAttempts int
	Attempt     int

	TraceParent string
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

	TraceParent string
}

type NodeFailedHandler func(ctx context.Context, payload NodeFailedPayload) error
type NodeFailedPayload struct {
	RunID       string
	NodeID      string
	Attempt     int
	MaxAttempts int
	Error       string
	Logs        []models.LogRecord

	TraceParent string
}
