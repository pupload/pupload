package flows

import (
	"encoding/json"
	"fmt"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
)

func (f *FlowService) ExecuteNode(n *models.Node) error {
	defName := n.DefName
	fmt.Println(n)

	nodeDef, exists := f.NodeDefs[defName]

	if !exists {
		return fmt.Errorf("Node Defintion %s does not exist.", defName)
	}

	task, err := NewNodeExecuteTask(nodeDef, *n)

	if err != nil {
		return err
	}
	f.AsynqClient.Enqueue(task)

	return nil
}

func NewNodeExecuteTask(nodeDef models.NodeDef, node models.Node) (*asynq.Task, error) {
	payload, err := json.Marshal(models.NodeExecutePayload{
		NodeDef: nodeDef,
		Node:    node,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeNodeExecute, payload), nil
}
