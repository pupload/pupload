package validation

import (
	"errors"
	"fmt"
	"slices"

	mimetypes "github.com/pupload/pupload/internal/mimetype"
	"github.com/pupload/pupload/internal/models"
)

func ValidateFlow(flow models.Flow, nodeDefs []models.NodeDef) error {

	edgeCount := generateEdgeCount(flow)

	// Check that all node definitions exist

	// Check that all edges are valid

	// Check for cycles

	if !isFlowDAG(flow) {
		return fmt.Errorf("Flow %s contains cycles", flow.Name)
	}

	// Check that all stores are valid

	// Check that all data wells are valid

	dwEdgeSeen := map[string]struct{}{}
	for _, well := range flow.DataWells {
		if !isValidDatawellSource(well) {
			return fmt.Errorf("Datawell with edge %s: %s is an invalid source type", well.Edge, *well.Source)
		}

		if _, ok := edgeCount[well.Edge]; !ok {
			return fmt.Errorf("Datawell with edge %s: edge %s is not defined in nodes", well.Edge, well.Edge)
		}

		if _, seen := dwEdgeSeen[well.Edge]; seen {
			return fmt.Errorf("Datawell with edge %s: two datawells must not share the same edge", well.Edge)
		}

		dwEdgeSeen[well.Edge] = struct{}{}
	}

	return nil
}

var ErrNodeTypeMismatch = errors.New("Invalid type on node")
var ErrNodeDefNotFound = errors.New("Node def not found")

func typeCheckFlow(flow models.Flow, nodeDefs []models.NodeDef) error {

	// edgeTypeMap := make(map[string]string)

	type EdgeNodeKey [2]string

	inputSet := make(map[EdgeNodeKey]mimetypes.MimeSet)
	outputSet := make(map[string]mimetypes.MimeSet)

	for _, node := range flow.Nodes {

		var def *models.NodeDef

		for _, d := range nodeDefs {
			if d.Name == node.Uses {
				def = &d
			}
		}

		if def == nil {
			return ErrNodeDefNotFound
		}

		for _, inEdge := range node.Inputs {
			var edgeDef models.NodeEdgeDef
			for _, ed := range def.Inputs {
				if ed.Name == inEdge.Name {
					edgeDef = ed
				}
			}

			set, err := mimetypes.CreateMimeSet(edgeDef.Type)
			if err != nil {
				return fmt.Errorf("Error type checking node %s on flow %s: %w", node.ID, flow.Name, err)
			}

			inputSet[EdgeNodeKey{inEdge.Edge, node.ID}] = *set
		}

		for _, outEdge := range node.Outputs {
			var edgeDef models.NodeEdgeDef
			for _, ed := range def.Outputs {
				if ed.Name == outEdge.Name {
					edgeDef = ed
				}
			}

			set, err := mimetypes.CreateMimeSet(edgeDef.Type)
			if err != nil {
				return fmt.Errorf("Error type checking node %s on flow %s: %w", node.ID, flow.Name, err)
			}

			outputSet[outEdge.Edge] = *set
		}
	}

	for _, node := range flow.Nodes {
		for _, in := range node.Inputs {
			inputTypes := inputSet[EdgeNodeKey{in.Edge, node.ID}]
			outputTypes := outputSet[in.Edge]

			intersection := inputTypes.Intersection(outputTypes)
			if intersection.IsEmpty() {
				return fmt.Errorf("no type overlap between input %s and output on edge %s", in.Name, in.Edge)
			}
		}
	}

	return nil

}

func doTypeSetsOverlap() {

}

func isValidDatawellSource(dw models.DataWell) bool {
	allowed_well_sources := []string{"upload", "static", "webhook"}

	if dw.Source == nil {
		return true
	}

	if !slices.Contains(allowed_well_sources, *dw.Source) {
		return false
	}

	return true
}

func isFlowDAG(flow models.Flow) bool {

	edgeProducers := make(map[string]string)
	edgeConsumers := make(map[string][]string)

	for _, node := range flow.Nodes {
		for _, output := range node.Outputs {
			edgeProducers[output.Edge] = node.ID
		}

		for _, input := range node.Inputs {
			edgeConsumers[input.Edge] = append(edgeConsumers[input.Edge], node.ID)
		}
	}

	adjacencyList := make(map[string][]string)
	inDegree := make(map[string]int)

	for edgeName, producerID := range edgeProducers {
		consumers := edgeConsumers[edgeName]
		for _, consumerID := range consumers {
			adjacencyList[producerID] = append(adjacencyList[producerID], consumerID)
			inDegree[consumerID]++
		}
	}

	q := make([]string, 0)
	for _, node := range flow.Nodes {
		if inDegree[node.ID] == 0 {
			q = append(q, node.ID)
		}
	}

	processedCount := 0

	for len(q) > 0 {
		current := q[0]
		q = q[1:]
		processedCount++

		for _, successor := range adjacencyList[current] {
			inDegree[successor]--
			if inDegree[successor] == 0 {
				q = append(q, successor)
			}
		}

	}

	return processedCount == len(flow.Nodes)
}

type EdgeCount map[string]int

func generateEdgeCount(flow models.Flow) EdgeCount {

	set := make(EdgeCount)

	for _, node := range flow.Nodes {
		for _, input := range node.Inputs {
			set[input.Edge] += 1
		}

		for _, output := range node.Outputs {
			set[output.Edge] += 1
		}
	}

	return set
}
