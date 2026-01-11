package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"pupload/internal/logging"
	"pupload/internal/models"
	"pupload/internal/stores"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"
	"time"

	"github.com/google/uuid"
)

type RuntimeFlow struct {
	Flow     models.Flow
	FlowRun  models.FlowRun
	NodeDefs []models.NodeDef

	nodes  map[string]RuntimeNode
	stores map[string]models.Store

	log *slog.Logger

	TraceParent string
}

type RuntimeNode struct {
	*models.Node
	NodeDef models.NodeDef
}

func CreateRuntimeFlow(ctx context.Context, flow models.Flow, nodeDefs []models.NodeDef) (RuntimeFlow, error) {
	// Unmarshal Stores

	runtimeFlow := RuntimeFlow{
		Flow:     flow,
		NodeDefs: nodeDefs,

		stores: make(map[string]models.Store),
		nodes:  make(map[string]RuntimeNode),

		TraceParent: telemetry.InjectContext(ctx),
	}

	runtimeFlow.constructLogger()
	runtimeFlow.constructStores()
	err := runtimeFlow.constructRuntimeNode()
	if err != nil {
		return runtimeFlow, err
	}

	runtimeFlow.createFlowRun()

	if err := runtimeFlow.initialDatawellInput(); err != nil {
		return runtimeFlow, err
	}

	return runtimeFlow, nil
}

func (rt *RuntimeFlow) RebuildRuntimeFlow() {

	rt.nodes = make(map[string]RuntimeNode)
	rt.stores = make(map[string]models.Store)

	rt.constructLogger()
	rt.constructStores()
	rt.constructRuntimeNode()
}

func (rt *RuntimeFlow) createFlowRun() {

	id := uuid.Must(uuid.NewV7())

	waitingUrls := make([]models.WaitingURL, 0)
	artifacts := make(map[string]models.Artifact)
	nodeStates := make(map[string]models.NodeState)

	for _, node := range rt.nodes {
		nodeStates[node.ID] = models.NodeState{Status: models.NODERUN_IDLE, Logs: []models.LogRecord{}}
	}

	value := models.FlowRun{
		ID:          id.String(),
		NodeState:   nodeStates,
		Status:      models.FLOWRUN_STOPPED,
		WaitingURLs: waitingUrls,
		Artifacts:   artifacts,
	}

	rt.FlowRun = value
}

func (rt *RuntimeFlow) constructRuntimeNode() error {
	for _, node := range rt.Flow.Nodes {
		found := false
		defName := node.DefName

		for _, def := range rt.NodeDefs {
			if defName == fmt.Sprintf("%s/%s", def.Publisher, def.Name) {
				found = true
				rt.nodes[node.ID] = RuntimeNode{Node: &node, NodeDef: def}
				break
			}
		}

		if !found {
			return fmt.Errorf("unable to find node with defName %s", defName)
		}
	}

	return nil
}

func (rt *RuntimeFlow) constructStores() {
	for _, storeInput := range rt.Flow.Stores {
		store, err := stores.UnmarshalStore(storeInput)
		if err != nil {
			rt.log.Warn("Invalid store definition", "flow", rt.Flow.Name, "store", storeInput.Name, "error", err.Error())
			continue
		}

		rt.stores[storeInput.Name] = store
	}
}

func (rt *RuntimeFlow) constructLogger() {
	rt.log = logging.ForService("flow_runtime").With("run_id", rt.FlowRun.ID)
}

func (rt *RuntimeFlow) initialDatawellInput() error {
	for _, dw := range rt.Flow.DataWells {

		var err error

		if dw.Source == nil {
			continue
		}

		switch *dw.Source {
		case "upload":
			err = rt.handleUploadDatawell(dw)

		case "static":
			// TODO: implement static resource grab
		case "webhook":
			// TODO: implement webhook data source
		}

		if err != nil {
			rt.log.Error("error processing datawells", "err", err)
			rt.FlowRun.Status = models.FLOWRUN_ERROR
			return err
		}
	}

	return nil
}

func (rt *RuntimeFlow) handleUploadDatawell(dw models.DataWell) error {

	key := rt.processDatawellKey(dw)
	store, ok := rt.stores[dw.Store]
	fmt.Println(store)
	if !ok {
		return fmt.Errorf("store %s does not exist", dw.Store)
	}

	artifact := models.Artifact{
		StoreName:  dw.Store,
		ObjectName: key,
		EdgeName:   dw.Edge,
	}

	url, err := store.PutURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
	if err != nil {
		return err
	}

	waitingURL := models.WaitingURL{
		Artifact: artifact,
		PutURL:   url.String(),
		TTL:      time.Now().Add(1 * time.Hour),
	}

	rt.FlowRun.WaitingURLs = append(rt.FlowRun.WaitingURLs, waitingURL)
	return nil
}

func (rt *RuntimeFlow) processDatawellKey(dw models.DataWell) string {
	if dw.Key == nil {
		return fmt.Sprintf("%s_%s", dw.Edge, rt.FlowRun.ID)
	}

	// TODO:
	return "TODOTESTPLEASEFIX"
}

// Returns nil if flow is already running.
// Returns error if flow is in error state or already complete
func (rt *RuntimeFlow) Start(s syncplane.SyncLayer) error {
	switch rt.FlowRun.Status {

	case models.FLOWRUN_STOPPED:
		rt.FlowRun.Status = models.FLOWRUN_WAITING

	case models.FLOWRUN_COMPLETE:
		return fmt.Errorf("runtime already complete")

	case models.FLOWRUN_ERROR:
		return fmt.Errorf("runtime in error state")

	case models.FLOWRUN_WAITING:

	case models.FLOWRUN_RUNNING:

	default:
		return fmt.Errorf("no case")
	}

	rt.Step(s)
	return nil
}
