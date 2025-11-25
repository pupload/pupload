package flows

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FlowRunStatus string

const (
	FLOWRUN_IDLE     FlowRunStatus = "IDLE"
	FLOWRUN_WAITING  FlowRunStatus = "WAITING"
	FLOWRUN_RUNNING  FlowRunStatus = "RUNNING"
	FLOWRUN_STOPPED  FlowRunStatus = "STOPPED"
	FLOWRUN_COMPLETE FlowRunStatus = "COMPLETE"
	FLOWRUN_ERROR    FlowRunStatus = "ERROR"
)

type Artifact struct {
	StoreName  string
	ObjectName string
	EdgeName   string
}

type WaitingURL struct {
	Artifact Artifact
	PutURL   string
}

type FlowRun struct {
	ID          string
	FlowName    string
	NodesLeft   map[int]struct{}
	Status      FlowRunStatus
	Artifacts   map[string]Artifact // Maps given edge ID to Artifact
	WaitingURLs []WaitingURL
}

func (f *FlowService) CreateFlowRun(flowName string) (FlowRun, error) {

	id := uuid.Must(uuid.NewV7())
	key := fmt.Sprintf("flowrun:%s", id)

	nodeCount := len(f.FlowList[flowName].Nodes)
	nodesLeft := make(map[int]struct{})
	for i := range nodeCount {
		nodesLeft[i] = struct{}{}
	}

	waitingUrls := make([]WaitingURL, 0)
	artifacts := make(map[string]Artifact)

	value := FlowRun{
		ID:          id.String(),
		FlowName:    flowName,
		NodesLeft:   nodesLeft,
		Status:      FLOWRUN_IDLE,
		WaitingURLs: waitingUrls,
		Artifacts:   artifacts,
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

func (f *FlowService) updateFlowRun(flowRun FlowRun) error {
	key := fmt.Sprintf("flowrun:%s", flowRun.ID)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(flowRun); err != nil {
		return err
	}

	if err := f.RedisClient.Set(context.TODO(), key, buf.Bytes(), 0).Err(); err != nil {
		return err
	}

	return nil

}

func (f *FlowService) initalizeWaitingURLs(flowrun *FlowRun) error {

	flow, err := f.GetFlow(flowrun.FlowName)
	if err != nil {
		return err
	}

	for _, datawell := range flow.DataWells {
		if datawell.Type != "dynamic" {
			continue
		}

		var key string
		if datawell.Key == nil {
			key = fmt.Sprintf("%s_%s", datawell.Edge, flowrun.ID)
		} else {
			key = f.getDataWellKey("beepboop", *flowrun)
		}

		store, ok := f.GetStore(flowrun.FlowName, datawell.Store)
		if !ok {
			return fmt.Errorf("Store %s does not exists", datawell.Store)
		}

		artifact := Artifact{
			StoreName:  datawell.Store,
			ObjectName: key,
			EdgeName:   datawell.Edge,
		}

		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
		if err != nil {
			return fmt.Errorf("Store %s could not generate put url: %s", datawell.Store, err)
		}

		WaitingURL := WaitingURL{
			Artifact: artifact,
			PutURL:   url.String(),
		}

		flowrun.WaitingURLs = append(flowrun.WaitingURLs, WaitingURL)

	}

	err = f.updateFlowRun(*flowrun)
	if err != nil {
		return err
	}

	return nil
}

func (f *FlowService) getDataWellKey(key string, flowrun FlowRun) string {
	return key
}

// func (f *FlowService) initalizeWaitingURLs(flowrun *FlowRun) error {

// 	for nodeID := range f.NodeLength(flowrun.FlowName) {
// 		node := f.GetNode(flowrun.FlowName, nodeID)
// 		for _, input := range node.Inputs {
// 			if input.Store == nil {
// 				continue
// 			}

// 			store, ok := f.GetStore(flowrun.FlowName, *input.Store)
// 			if !ok {
// 				continue
// 			}

// 			artifact := Artifact{
// 				StoreName:  *input.Store,
// 				ObjectName: input.Edge,
// 			}

// 			url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
// 			if err != nil {
// 				fmt.Println(err)
// 				continue
// 			}

// 			waitingURL := WaitingURL{
// 				Artifact: artifact,
// 				PutURL:   url.String(),
// 			}

// 			flowrun.WaitingURLs = append(flowrun.WaitingURLs, waitingURL)

// 		}
// 	}

// 	err := f.updateFlowRun(*flowrun)
// 	if err != nil {
// 		fmt.Println(err)f
// 		return err
// 	}

// 	return nil
// }

func (f *FlowService) checkWaitingUrls(flowrun *FlowRun) (bool, error) {

	fileExists := false

	for i, waitingUrl := range flowrun.WaitingURLs {
		store, ok := f.GetStore(flowrun.FlowName, waitingUrl.Artifact.StoreName)
		if !ok {
			return false, fmt.Errorf("store referenced in waitingURL does not exists. this should never happen")
		}

		exists := store.Exists(waitingUrl.Artifact.ObjectName)
		if exists {

			flowrun.Artifacts[waitingUrl.Artifact.EdgeName] = waitingUrl.Artifact

			// TODO: Handle oob's edge case
			flowrun.WaitingURLs = append(flowrun.WaitingURLs[:i], flowrun.WaitingURLs[i+1:]...)
			fileExists = true
		}
	}

	return fileExists, nil
}
