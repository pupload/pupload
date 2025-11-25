package models

const (
	TypeFlowStep        = "flow:step"
	TypeNodeExecute     = "node:execute"
	TypeNodeFinished    = "node:finished"
	TypeControllerClean = "controller:clean"
)

type FlowStepPayload struct {
	RunID string
}

type NodeExecutePayload struct {
	NodeDef    NodeDef
	Node       Node
	InputURLs  map[string]string
	OutputURLs map[string]string
}
