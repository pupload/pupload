package models

import (
	"time"
)

type FlowRunStatus string

const (
	FLOWRUN_STOPPED  FlowRunStatus = "STOPPED"
	FLOWRUN_WAITING  FlowRunStatus = "WAITING"
	FLOWRUN_RUNNING  FlowRunStatus = "RUNNING"
	FLOWRUN_COMPLETE FlowRunStatus = "COMPLETE"
	FLOWRUN_ERROR    FlowRunStatus = "ERROR"
)

type NodeRunStatus string

const (
	NODERUN_IDLE     NodeRunStatus = "IDLE"
	NODERUN_READY    NodeRunStatus = "READY"
	NODERUN_RUNNING  NodeRunStatus = "RUNNING"
	NODERUN_RETRYING NodeRunStatus = "RETRYING"
	NODERUN_COMPLETE NodeRunStatus = "COMPLETE"
	NODERUN_ERROR    NodeRunStatus = "ERROR"
)

type NodeState struct {
	Status     NodeRunStatus
	Logs       []LogRecord
	Error      string
	RetryCount int
	MaxRetries int
}

type Artifact struct {
	StoreName  string
	ObjectName string
	EdgeName   string
}

type WaitingURL struct {
	Artifact Artifact
	PutURL   string
	TTL      time.Time
}

type FlowRun struct {
	ID string

	NodeState   map[string]NodeState
	Status      FlowRunStatus
	Artifacts   map[string]Artifact // Maps given edge ID to Artifact
	WaitingURLs []WaitingURL
}
