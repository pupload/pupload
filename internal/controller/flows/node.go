package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pupload/internal/models"
	"time"

	"github.com/hibiken/asynq"
)

func (f *FlowService) GetNode(flowName string, nodeIndex int) models.Node {
	return f.FlowList[flowName].Nodes[nodeIndex]

}

func (f *FlowService) HandleExecuteNode(run FlowRun, nodeIndex int) {

	node := f.GetNode(run.FlowName, nodeIndex)

	nodeDef, exists := f.NodeDefs[node.DefName]
	if !exists {
		log.Fatalf("Attempting to run node that does not have definition")
	}

	inputs := make(map[string]string)
	for _, edge := range node.Inputs {
		artifact := run.Artifacts[edge.ID]

		store, _ := f.GetStore(run.FlowName, artifact.StoreName)
		url, err := store.GetURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
		if err != nil {
			fmt.Printf(err.Error())
		}

		inputs[edge.Name] = url.String()
	}

	f.ExecuteNode(&nodeDef, &node, inputs, nil)

	delete(run.NodesLeft, nodeIndex)
	f.updateFlowRun(run)
}

func (f *FlowService) ExecuteNode(nodeDef *models.NodeDef, n *models.Node, inputURLs map[string]string, outputURLs map[string]string) error {

	task, err := NewNodeExecuteTask(*nodeDef, *n, inputURLs, outputURLs)

	if err != nil {
		return err
	}
	f.AsynqClient.Enqueue(task, asynq.Queue("worker"))

	return nil
}

func NewNodeExecuteTask(nodeDef models.NodeDef, node models.Node, inputURLs map[string]string, outputURLs map[string]string) (*asynq.Task, error) {
	payload, err := json.Marshal(models.NodeExecutePayload{
		NodeDef:    nodeDef,
		Node:       node,
		InputURLs:  inputURLs,
		OutputURLs: outputURLs,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeNodeExecute, payload), nil
}

func HandleNodeFinishedTask(ctx context.Context, t *asynq.Task) error {
	return nil
}
