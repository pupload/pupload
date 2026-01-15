package node

import (
	"context"
	"log/slog"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"

	"github.com/moby/moby/api/types/container"
)

func (ns *NodeService) FinishedMiddleware(ctx context.Context, payload syncplane.NodeExecutePayload, resource container.Resources) error {

	ctx = telemetry.ExtractContext(ctx, payload.TraceParent)

	logs := make([]models.LogRecord, 0, 64)
	ch := &logging.CollectHandler{
		Inner:   logging.Root().Handler(),
		Records: &logs,
	}

	jobLog := slog.New(ch)
	jobLog.With(
		"run_id", payload.RunID,
		"node_id", payload.Node.ID,
		"nodedef_publisher", payload.NodeDef.Publisher,
		"nodedef_name", payload.NodeDef.Name,
		"container_image", payload.NodeDef.Image,
	)

	ctx = logging.CtxWithLogger(ctx, jobLog)

	err := ns.NodeExecute(ctx, payload, resource)
	if err == nil {

		if err := ns.SyncLayer.EnqueueNodeFinished(syncplane.NodeFinishedPayload{
			RunID:  payload.RunID,
			NodeID: payload.Node.ID,
			Logs:   logs,
		}); err != nil {

		}
	}

	if err != nil {
		jobLog.Error(err.Error())
	}

	return err
}
