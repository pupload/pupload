package flows

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/google/uuid"
)

type FlowRunStatus string

const (
	FLOWRUN_IDLE FlowRunStatus = "IDLE"
)

type Artifact struct {
	Store      string
	ObjectName string
}

type FlowRun struct {
	ID          string
	FlowName    string
	CurrentNode int
	Status      FlowRunStatus
	Artifacts   map[string]Artifact // Maps given edge ID to
}

func (f *FlowService) CreateFlowRun(flowName string) (FlowRun, error) {

	id := uuid.Must(uuid.NewV7())

	key := fmt.Sprintf("flowrun:%s", id)
	value := FlowRun{
		ID:          id.String(),
		FlowName:    flowName,
		CurrentNode: 0,
		Status:      FLOWRUN_IDLE,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return value, err
	}

	if err := f.RedisClient.Set(context.TODO(), key, buf.Bytes(), 0).Err(); err != nil {
		return value, err
	}

	return value, nil

}

func (f *FlowService) GetFlowRun(id string) (FlowRun, error) {
	key := fmt.Sprintf("flowrun:%s", id)
	raw, err := f.RedisClient.Get(context.TODO(), key).Bytes()
	if err != nil {
		return FlowRun{}, fmt.Errorf("FlowRun %s does not exist", id)
	}

	var val FlowRun
	dec := gob.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&val); err != nil {
		return FlowRun{}, err
	}

	return val, nil

}

/*
func (f *FlowService) updateFlowRun(id string) {
	key := fmt.Sprintf("flowrun:%s", id)

}
*/
