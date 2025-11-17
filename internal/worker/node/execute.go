package node

import (
	"context"
	"encoding/json"
	"fmt"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
)

func HandleNodeExecuteTask(ctx context.Context, t *asynq.Task) error {
	var p models.NodeExecutePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	fmt.Println("YAYYY")

	return nil
}
