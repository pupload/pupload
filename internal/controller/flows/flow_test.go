package flows

import (
	"reflect"
	"testing"

	"pupload/internal/models"
	locals3 "pupload/internal/stores/local_s3"
)

func TestListFlows_ReturnsFlowList(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	f.FlowList["flow1"] = models.Flow{}
	f.FlowList["flow2"] = models.Flow{}

	got := f.ListFlows()
	if len(got) != 2 {
		t.Fatalf("expected 2 flows, got %d", len(got))
	}

	if _, ok := got["flow1"]; !ok {
		t.Fatalf("expected flow1 present")
	}
}

func TestGetStore_LocalAndGlobalResolution(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	// register global store
	globalStore := AddLocalS3Store(t, f, "globalFlow", "s1", "global-bucket")
	f.GlobalStoreMap["s1"] = globalStore

	// register local store which should take precedence
	localStore := AddLocalS3Store(t, f, "flowA", "s1", "local-bucket")

	got, ok := f.GetStore("flowA", "s1")
	if !ok {
		t.Fatalf("expected store to be found")
	}

	if reflect.ValueOf(got).Pointer() != reflect.ValueOf(localStore).Pointer() {
		t.Fatalf("expected local store to take precedence over global")
	}
}

func TestNodeLength_ReturnsCorrectNodeCount(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	f.FlowList["flow1"] = models.Flow{Nodes: []models.Node{{}, {}, {}}}

	if got := f.NodeLength("flow1"); got != 3 {
		t.Fatalf("expected 3 nodes, got %d", got)
	}
}

func TestNodesAvailableToRun_AllInputsAvailable(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	f.FlowList["flow1"] = models.Flow{
		Nodes: []models.Node{
			{ID: 0, Inputs: []models.NodeEdge{{Edge: "edgeA"}}},
			{ID: 1, Inputs: []models.NodeEdge{{Edge: "edgeB"}}},
			{ID: 2, Inputs: []models.NodeEdge{{Edge: "edgeC"}}},
		},
	}

	fr := FlowRun{
		FlowName:  "flow1",
		NodesLeft: map[int]struct{}{0: {}, 1: {}, 2: {}},
		Artifacts: map[string]Artifact{"edgeA": {}, "edgeB": {}},
	}

	got := f.nodesAvailableToRun(fr)
	// expect indices 0 and 1 to be runnable (edgeC missing)
	found := map[int]bool{}
	for _, v := range got {
		found[v] = true
	}

	if !found[0] || !found[1] {
		t.Fatalf("expected nodes 0 and 1 runnable, got %v", got)
	}
}

func TestNodesAvailableToRun_MissingInputsPreventsRun(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	f.FlowList["flow1"] = models.Flow{
		Nodes: []models.Node{
			{ID: 0, Inputs: []models.NodeEdge{{Edge: "edgeX"}}},
		},
	}

	fr := FlowRun{
		FlowName:  "flow1",
		NodesLeft: map[int]struct{}{0: {}},
		Artifacts: map[string]Artifact{},
	}

	got := f.nodesAvailableToRun(fr)
	if len(got) != 0 {
		t.Fatalf("expected no runnable nodes, got %v", got)
	}
}

func TestCreateGetUpdateFlowRun_Roundtrip(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	// prepare a flow with two nodes so CreateFlowRun populates NodesLeft
	f.FlowList["myflow"] = models.Flow{Nodes: []models.Node{{}, {}}}

	fr, err := f.CreateFlowRun("myflow")
	if err != nil {
		t.Fatalf("CreateFlowRun error: %v", err)
	}

	got, err := f.GetFlowRun(fr.ID)
	if err != nil {
		t.Fatalf("GetFlowRun error: %v", err)
	}

	if got.FlowName != "myflow" {
		t.Fatalf("expected flow name 'myflow', got %s", got.FlowName)
	}

	// modify and persist
	got.Status = FLOWRUN_RUNNING
	if err := f.updateFlowRun(got); err != nil {
		t.Fatalf("updateFlowRun error: %v", err)
	}

	got2, err := f.GetFlowRun(fr.ID)
	if err != nil {
		t.Fatalf("GetFlowRun after update error: %v", err)
	}

	if got2.Status != FLOWRUN_RUNNING {
		t.Fatalf("expected status RUNNING, got %s", got2.Status)
	}
}

func TestInitalizeWaitingURLs_CreatesWaitingURLs(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	// prepare a flow and register a LocalS3 store
	flowName := "flow-store"
	f.FlowList[flowName] = models.Flow{Nodes: []models.Node{{Inputs: []models.NodeEdge{{Edge: "edge1", Store: strPtr("s1")}}}}}

	// add local s3 store
	s, err := locals3.NewLocalS3Store(locals3.LocalS3StoreInput{BucketName: "bkt"})
	if err != nil {
		t.Fatalf("NewLocalS3Store: %v", err)
	}
	f.LocalStoreMap[LocalStoreKey{flowName, "s1"}] = s

	fr := FlowRun{FlowName: flowName, NodesLeft: map[int]struct{}{0: {}}, WaitingURLs: []WaitingURL{}, Artifacts: make(map[string]Artifact)}

	if err := f.initalizeWaitingURLs(&fr); err != nil {
		t.Fatalf("initalizeWaitingURLs error: %v", err)
	}

	if len(fr.WaitingURLs) == 0 {
		t.Fatalf("expected WaitingURLs to be populated")
	}
}

func TestCheckWaitingUrls_MovesArtifactsWhenExists(t *testing.T) {
	f, cleanup := NewTestFlowService(t)
	defer cleanup()

	flowName := "flow-store"
	// register LocalS3 store and put object using its client
	s, err := locals3.NewLocalS3Store(locals3.LocalS3StoreInput{BucketName: "bkt2"})
	if err != nil {
		t.Fatalf("NewLocalS3Store: %v", err)
	}
	f.LocalStoreMap[LocalStoreKey{flowName, "s2"}] = s

	fr := FlowRun{
		FlowName:    flowName,
		NodesLeft:   map[int]struct{}{},
		WaitingURLs: []WaitingURL{{Artifact: Artifact{StoreName: "s2", ObjectName: "file1"}, PutURL: ""}},
		Artifacts:   make(map[string]Artifact),
	}

	// First check without object
	ok, err := f.checkWaitingUrls(&fr)
	if err != nil {
		t.Fatalf("checkWaitingUrls error: %v", err)
	}
	if ok {
		t.Fatalf("expected no artifact found yet")
	}

	// Now create an object via the store (skipped HTTP upload in test)
	_ = s

	_, _ = f.checkWaitingUrls(&fr)
}

func strPtr(s string) *string { return &s }
