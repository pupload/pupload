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

type FlowRun struct {
	ID          string
	FlowName    string
	CurrentNode int
	Status      FlowRunStatus
}

func (f *FlowService) createFlowRun(flowName string) (string, error) {

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
		return "", err
	}

	f.RedisClient.Set(context.TODO(), key, buf, 0).Result()

	return id.String(), nil

}

func (f *FlowService) getFlowRun(id string) (FlowRun, error) {
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
