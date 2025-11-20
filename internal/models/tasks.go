package models

const (
	TypeNodeExecute     = "node:execute"
	TypeNodeFinished    = "node:finished"
	TypeControllerClean = "controller:clean"
)

type NodeExecutePayload struct {
	NodeDef    NodeDef
	Node       Node
	InputURLs  map[string]string
	OutputURLs map[string]string
}
