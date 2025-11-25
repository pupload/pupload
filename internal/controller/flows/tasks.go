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

func (f *FlowService) AsynqConfigureHandlers() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(models.TypeFlowStep, f.HandleFlowStepTask)
	mux.HandleFunc(models.TypeNodeFinished, HandleNodeFinishedTask)

	return mux

}
func (f *FlowService) HandleFlowStepTask(ctx context.Context, t *asynq.Task) error {
	var p models.FlowStepPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("Error unmarshaling in Flow ID: %s\n", p.RunID)
		return err
	}

	log.Printf("Running Flow ID: %s\n", p.RunID)

	go f.HandleStepFlow(ctx, p.RunID)

	return nil
}

func NewFlowStepTask(runID string) (*asynq.Task, error) {
	payload, err := json.Marshal(models.FlowStepPayload{
		RunID: runID,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(models.TypeFlowStep, payload), nil
}

func (f *FlowService) EnqueueFlowStepTask(runID string, delay time.Duration) error {
	log.Printf("Attmpting Enqueue of Flow ID: %s\n", runID)
	task, err := NewFlowStepTask(runID)
	if err != nil {
		return fmt.Errorf("Could not enqueue flow step task: %s", err)
	}

	_, err = f.AsynqClient.Enqueue(task, asynq.ProcessIn(delay), asynq.Queue("controller"))
	if err != nil {
		return fmt.Errorf("Error enqueing flow step task: %s\n", err)
	}

	return err
}
