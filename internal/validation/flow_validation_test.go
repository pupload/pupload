package validation

import (
	"testing"

	"github.com/pupload/pupload/internal/models"
)

// Helper function to create string pointers
func ptr(s string) *string {
	return &s
}

// newTestFlow creates a valid base flow for testing
func newTestFlow(t *testing.T, options ...func(*models.Flow)) models.Flow {
	t.Helper()

	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge1"}},
			},
		},
		DataWells: []models.DataWell{},
		Stores:    []models.StoreInput{{Name: "test-store", Type: "s3"}},
	}

	// Apply optional modifications
	for _, opt := range options {
		opt(&flow)
	}

	return flow
}

// newTestNodeDef creates a test node definition
func newTestNodeDef(t *testing.T, name string) models.NodeDef {
	t.Helper()

	return models.NodeDef{
		ID:        1,
		Publisher: "test",
		Name:      name,
		Image:     "test:latest",
		Inputs:    []models.NodeEdgeDef{{Name: "input", Type: []models.MimeType{"string"}, Required: false}},
		Outputs:   []models.NodeEdgeDef{{Name: "output", Type: []models.MimeType{"string"}, Required: false}},
	}
}

// newTestDataWell creates a test data well
func newTestDataWell(t *testing.T, edge string, source *string) models.DataWell {
	t.Helper()

	return models.DataWell{
		Edge:   edge,
		Store:  "test-store",
		Source: source,
	}
}

// newTestStore creates a test store input
func newTestStore(t *testing.T, name string) models.StoreInput {
	t.Helper()

	return models.StoreInput{
		Name: name,
		Type: "s3",
	}
}

// ============================================================================
// ValidateFlow - Implemented functionality tests
// ============================================================================

func TestValidateFlow_ValidFlow(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("upload")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("ValidateFlow() error = %v, want nil", err)
	}
}

func TestValidateFlow_InvalidDatawellSource(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("invalid-source")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for invalid datawell source, got nil")
	}
}

func TestValidateFlow_DatawellEdgeNotInNodes(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "nonexistent-edge", ptr("upload")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for datawell edge not in nodes, got nil")
	}
}

func TestValidateFlow_DuplicateDatawellEdges(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("upload")),
			newTestDataWell(t, "edge1", ptr("static")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for duplicate datawell edges, got nil")
	}
}

// ============================================================================
// ValidateFlow - DefaultDataWell validation tests
// ============================================================================

func TestValidateFlow_MissingDefaultDataWellWithUncoveredEdges(t *testing.T) {
	t.Skip("DefaultDataWell validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		// Edge exists but no DataWell for it, and DefaultDataWell is nil
		f.DataWells = []models.DataWell{}
		f.DefaultDataWell = nil
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for missing DefaultDataWell with uncovered edges, got nil")
	}
}

func TestValidateFlow_ValidDefaultDataWell(t *testing.T) {
	t.Skip("DefaultDataWell validation not yet implemented")

	defaultDataWell := models.DataWell{
		Edge:   "",
		Store:  "test-store",
		Source: ptr("upload"),
	}

	flow := newTestFlow(t, func(f *models.Flow) {
		f.DataWells = []models.DataWell{}
		f.DefaultDataWell = &defaultDataWell
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("ValidateFlow() error = %v, want nil (DefaultDataWell is set)", err)
	}
}

func TestValidateFlow_NoDefaultDataWellNeededWhenAllEdgesCovered(t *testing.T) {
	t.Skip("DefaultDataWell validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		// All edges have explicit DataWells
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("upload")),
		}
		f.DefaultDataWell = nil
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("ValidateFlow() error = %v, want nil (all edges covered)", err)
	}
}

// ============================================================================
// ValidateFlow - Unimplemented functionality tests
// ============================================================================

func TestValidateFlow_NodeDefNotFound(t *testing.T) {
	t.Skip("Node definition validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		f.Nodes = []models.Node{
			{
				ID:      "node1",
				Uses:    "nonexistent-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge1"}},
			},
		}
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("upload")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for node def not found, got nil")
	}
}

func TestValidateFlow_InvalidEdge_NonexistentSourceNode(t *testing.T) {
	t.Skip("Edge validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		f.Nodes = []models.Node{
			{
				ID:     "node1",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "edge-from-nowhere"}},
			},
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for edge with nonexistent source node, got nil")
	}
}

func TestValidateFlow_InvalidEdge_NonexistentTargetNode(t *testing.T) {
	t.Skip("Edge validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		f.Nodes = []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge-to-nowhere"}},
			},
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for edge with nonexistent target node, got nil")
	}
}

func TestValidateFlow_CycleDetection(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		// A -> B -> C -> A
		f.Nodes = []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "edge-C-A"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge-A-B"}},
			},
			{
				ID:      "B",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "edge-A-B"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge-B-C"}},
			},
			{
				ID:      "C",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "edge-B-C"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge-C-A"}},
			},
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for cycle in flow, got nil")
	}
}

func TestValidateFlow_SelfReferenceCycle(t *testing.T) {
	flow := newTestFlow(t, func(f *models.Flow) {
		// Node A references itself: A -> A
		f.Nodes = []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "self-loop"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "self-loop"}},
			},
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for self-reference cycle, got nil")
	}
}

func TestValidateFlow_InvalidStoreReference(t *testing.T) {
	t.Skip("Store reference validation not yet implemented")

	flow := newTestFlow(t, func(f *models.Flow) {
		f.Nodes = []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge1", Store: ptr("nonexistent-store")}},
			},
		}
		f.DataWells = []models.DataWell{
			newTestDataWell(t, "edge1", ptr("upload")),
		}
	})

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := ValidateFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("ValidateFlow() expected error for invalid store reference, got nil")
	}
}

// ============================================================================
// isValidDatawellSource tests
// ============================================================================

func TestIsValidDatawellSource_NilSource(t *testing.T) {
	dw := models.DataWell{
		Edge:   "test-edge",
		Store:  "test-store",
		Source: nil,
	}

	result := isValidDatawellSource(dw)
	if !result {
		t.Fatal("isValidDatawellSource() = false, want true for nil source")
	}
}

func TestIsValidDatawellSource_ValidSources(t *testing.T) {
	validSources := []string{"upload", "static", "webhook"}

	for _, source := range validSources {
		dw := models.DataWell{
			Edge:   "test-edge",
			Store:  "test-store",
			Source: ptr(source),
		}

		result := isValidDatawellSource(dw)
		if !result {
			t.Fatalf("isValidDatawellSource() = false for source %q, want true", source)
		}
	}
}

func TestIsValidDatawellSource_InvalidSource(t *testing.T) {
	dw := models.DataWell{
		Edge:   "test-edge",
		Store:  "test-store",
		Source: ptr("invalid-source"),
	}

	result := isValidDatawellSource(dw)
	if result {
		t.Fatal("isValidDatawellSource() = true for invalid source, want false")
	}
}

// ============================================================================
// generateEdgeCount tests
// ============================================================================

func TestGenerateEdgeCount_EmptyFlow(t *testing.T) {
	flow := models.Flow{
		Name:  "empty-flow",
		Nodes: []models.Node{},
	}

	edgeCount := generateEdgeCount(flow)
	if len(edgeCount) != 0 {
		t.Fatalf("generateEdgeCount() returned %d edges, want 0 for empty flow", len(edgeCount))
	}
}

func TestGenerateEdgeCount_SingleNode(t *testing.T) {
	flow := models.Flow{
		Name: "single-node-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "input-edge"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "output-edge"}},
			},
		},
	}

	edgeCount := generateEdgeCount(flow)
	if edgeCount["input-edge"] != 1 {
		t.Fatalf("generateEdgeCount()[\"input-edge\"] = %d, want 1", edgeCount["input-edge"])
	}
	if edgeCount["output-edge"] != 1 {
		t.Fatalf("generateEdgeCount()[\"output-edge\"] = %d, want 1", edgeCount["output-edge"])
	}
}

func TestGenerateEdgeCount_MultipleNodesSharedEdges(t *testing.T) {
	flow := models.Flow{
		Name: "multi-node-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "shared-edge"}},
			},
			{
				ID:     "node2",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "shared-edge"}},
			},
		},
	}

	edgeCount := generateEdgeCount(flow)
	if edgeCount["shared-edge"] != 2 {
		t.Fatalf("generateEdgeCount()[\"shared-edge\"] = %d, want 2", edgeCount["shared-edge"])
	}
}

func TestGenerateEdgeCount_NodeWithOnlyOutputs(t *testing.T) {
	flow := models.Flow{
		Name: "output-only-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "edge1"}},
			},
		},
	}

	edgeCount := generateEdgeCount(flow)
	if edgeCount["edge1"] != 1 {
		t.Fatalf("generateEdgeCount()[\"edge1\"] = %d, want 1", edgeCount["edge1"])
	}
}

// ============================================================================
// typeCheckFlow tests
// ============================================================================

func TestTypeCheckFlow_NodeDefNotFound(t *testing.T) {
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:   "node1",
				Uses: "nonexistent-def",
			},
		},
	}

	nodeDefs := []models.NodeDef{
		newTestNodeDef(t, "test-def"),
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err != ErrNodeDefNotFound {
		t.Fatalf("typeCheckFlow() error = %v, want ErrNodeDefNotFound", err)
	}
}

func TestTypeCheckFlow_ValidTypeMatch(t *testing.T) {
	// Create a flow where node1 outputs to node2, with matching types
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "producer",
				Outputs: []models.NodeEdge{{Name: "output", Edge: "shared-edge"}},
			},
			{
				ID:     "node2",
				Uses:   "consumer",
				Inputs: []models.NodeEdge{{Name: "input", Edge: "shared-edge"}},
			},
		},
	}

	nodeDefs := []models.NodeDef{
		{
			ID:        1,
			Publisher: "test",
			Name:      "producer",
			Image:     "test:latest",
			Outputs:   []models.NodeEdgeDef{{Name: "output", Type: []models.MimeType{"text/plain"}}},
		},
		{
			ID:        2,
			Publisher: "test",
			Name:      "consumer",
			Image:     "test:latest",
			Inputs:    []models.NodeEdgeDef{{Name: "input", Type: []models.MimeType{"text/plain"}}},
		},
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("typeCheckFlow() error = %v, want nil for matching types", err)
	}
}

func TestTypeCheckFlow_TypeMismatch(t *testing.T) {
	// Create a flow where node1 outputs incompatible type to node2
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "producer",
				Outputs: []models.NodeEdge{{Name: "output", Edge: "shared-edge"}},
			},
			{
				ID:     "node2",
				Uses:   "consumer",
				Inputs: []models.NodeEdge{{Name: "input", Edge: "shared-edge"}},
			},
		},
	}

	nodeDefs := []models.NodeDef{
		{
			ID:        1,
			Publisher: "test",
			Name:      "producer",
			Image:     "test:latest",
			Outputs:   []models.NodeEdgeDef{{Name: "output", Type: []models.MimeType{"text/plain"}}},
		},
		{
			ID:        2,
			Publisher: "test",
			Name:      "consumer",
			Image:     "test:latest",
			Inputs:    []models.NodeEdgeDef{{Name: "input", Type: []models.MimeType{"application/json"}}},
		},
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err == nil {
		t.Fatal("typeCheckFlow() expected error for type mismatch, got nil")
	}
}

func TestTypeCheckFlow_WildcardTypeMatch(t *testing.T) {
	// Create a flow where producer outputs text/plain and consumer accepts text/*
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "producer",
				Outputs: []models.NodeEdge{{Name: "output", Edge: "shared-edge"}},
			},
			{
				ID:     "node2",
				Uses:   "consumer",
				Inputs: []models.NodeEdge{{Name: "input", Edge: "shared-edge"}},
			},
		},
	}

	nodeDefs := []models.NodeDef{
		{
			ID:        1,
			Publisher: "test",
			Name:      "producer",
			Image:     "test:latest",
			Outputs:   []models.NodeEdgeDef{{Name: "output", Type: []models.MimeType{"text/plain"}}},
		},
		{
			ID:        2,
			Publisher: "test",
			Name:      "consumer",
			Image:     "test:latest",
			Inputs:    []models.NodeEdgeDef{{Name: "input", Type: []models.MimeType{"text/*"}}},
		},
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("typeCheckFlow() error = %v, want nil for wildcard type match", err)
	}
}

func TestTypeCheckFlow_MultipleCompatibleTypes(t *testing.T) {
	// Create a flow where producer outputs multiple types, consumer accepts one of them
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "producer",
				Outputs: []models.NodeEdge{{Name: "output", Edge: "shared-edge"}},
			},
			{
				ID:     "node2",
				Uses:   "consumer",
				Inputs: []models.NodeEdge{{Name: "input", Edge: "shared-edge"}},
			},
		},
	}

	nodeDefs := []models.NodeDef{
		{
			ID:        1,
			Publisher: "test",
			Name:      "producer",
			Image:     "test:latest",
			Outputs:   []models.NodeEdgeDef{{Name: "output", Type: []models.MimeType{"text/plain", "application/json"}}},
		},
		{
			ID:        2,
			Publisher: "test",
			Name:      "consumer",
			Image:     "test:latest",
			Inputs:    []models.NodeEdgeDef{{Name: "input", Type: []models.MimeType{"application/json", "application/xml"}}},
		},
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("typeCheckFlow() error = %v, want nil for overlapping types", err)
	}
}

func TestTypeCheckFlow_EmptyFlow(t *testing.T) {
	// Empty flow should pass type checking
	flow := models.Flow{
		Name:  "empty-flow",
		Nodes: []models.Node{},
	}

	nodeDefs := []models.NodeDef{}

	err := typeCheckFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("typeCheckFlow() error = %v, want nil for empty flow", err)
	}
}

func TestTypeCheckFlow_NodeWithNoConnections(t *testing.T) {
	// Node with no inputs/outputs should pass type checking
	flow := models.Flow{
		Name: "test-flow",
		Nodes: []models.Node{
			{
				ID:      "node1",
				Uses:    "standalone",
				Inputs:  []models.NodeEdge{},
				Outputs: []models.NodeEdge{},
			},
		},
	}

	nodeDefs := []models.NodeDef{
		{
			ID:        1,
			Publisher: "test",
			Name:      "standalone",
			Image:     "test:latest",
			Inputs:    []models.NodeEdgeDef{},
			Outputs:   []models.NodeEdgeDef{},
		},
	}

	err := typeCheckFlow(flow, nodeDefs)
	if err != nil {
		t.Fatalf("typeCheckFlow() error = %v, want nil for node with no connections", err)
	}
}

// ============================================================================
// isFlowDAG tests
// ============================================================================

func TestIsFlowDAG_EmptyFlow(t *testing.T) {
	flow := models.Flow{
		Name:  "empty-flow",
		Nodes: []models.Node{},
	}

	result := isFlowDAG(flow)
	if !result {
		t.Fatal("isFlowDAG() = false for empty flow, want true")
	}
}

func TestIsFlowDAG_LinearFlow(t *testing.T) {
	// A -> B -> C
	flow := models.Flow{
		Name: "linear-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-B"}},
			},
			{
				ID:      "B",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-B"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "B-C"}},
			},
			{
				ID:     "C",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "B-C"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if !result {
		t.Fatal("isFlowDAG() = false for linear flow, want true")
	}
}

func TestIsFlowDAG_DiamondPattern(t *testing.T) {
	// A -> B, A -> C, B -> D, C -> D (diamond pattern - valid DAG)
	flow := models.Flow{
		Name: "diamond-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out1", Edge: "A-B"}, {Name: "out2", Edge: "A-C"}},
			},
			{
				ID:      "B",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-B"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "B-D"}},
			},
			{
				ID:      "C",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-C"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "C-D"}},
			},
			{
				ID:     "D",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in1", Edge: "B-D"}, {Name: "in2", Edge: "C-D"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if !result {
		t.Fatal("isFlowDAG() = false for diamond pattern, want true")
	}
}

func TestIsFlowDAG_SimpleCycle(t *testing.T) {
	// A -> B -> A (simple cycle)
	flow := models.Flow{
		Name: "cycle-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "B-A"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-B"}},
			},
			{
				ID:      "B",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-B"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "B-A"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if result {
		t.Fatal("isFlowDAG() = true for simple cycle, want false")
	}
}

func TestIsFlowDAG_ComplexCycle(t *testing.T) {
	// A -> B -> C -> A (complex cycle)
	flow := models.Flow{
		Name: "complex-cycle-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "C-A"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-B"}},
			},
			{
				ID:      "B",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-B"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "B-C"}},
			},
			{
				ID:      "C",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "B-C"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "C-A"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if result {
		t.Fatal("isFlowDAG() = true for complex cycle, want false")
	}
}

func TestIsFlowDAG_SelfReference(t *testing.T) {
	// A -> A (self-reference)
	flow := models.Flow{
		Name: "self-ref-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "A-A"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-A"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if result {
		t.Fatal("isFlowDAG() = true for self-reference, want false")
	}
}

func TestIsFlowDAG_DisconnectedSubgraphs(t *testing.T) {
	// Two separate DAGs: A -> B and C -> D
	flow := models.Flow{
		Name: "disconnected-flow",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-B"}},
			},
			{
				ID:     "B",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "A-B"}},
			},
			{
				ID:      "C",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "C-D"}},
			},
			{
				ID:     "D",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "C-D"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if !result {
		t.Fatal("isFlowDAG() = false for disconnected subgraphs, want true")
	}
}

func TestIsFlowDAG_DisconnectedWithCycle(t *testing.T) {
	// One valid DAG (A -> B) and one cycle (C -> D -> C)
	flow := models.Flow{
		Name: "disconnected-with-cycle",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Outputs: []models.NodeEdge{{Name: "out", Edge: "A-B"}},
			},
			{
				ID:     "B",
				Uses:   "test-def",
				Inputs: []models.NodeEdge{{Name: "in", Edge: "A-B"}},
			},
			{
				ID:      "C",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "D-C"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "C-D"}},
			},
			{
				ID:      "D",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{{Name: "in", Edge: "C-D"}},
				Outputs: []models.NodeEdge{{Name: "out", Edge: "D-C"}},
			},
		},
	}

	result := isFlowDAG(flow)
	if result {
		t.Fatal("isFlowDAG() = true for disconnected graphs with cycle, want false")
	}
}

func TestIsFlowDAG_SingleNodeNoEdges(t *testing.T) {
	// Single isolated node with no connections
	flow := models.Flow{
		Name: "single-node",
		Nodes: []models.Node{
			{
				ID:      "A",
				Uses:    "test-def",
				Inputs:  []models.NodeEdge{},
				Outputs: []models.NodeEdge{},
			},
		},
	}

	result := isFlowDAG(flow)
	if !result {
		t.Fatal("isFlowDAG() = false for single node with no edges, want true")
	}
}
