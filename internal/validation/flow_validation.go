package validation

import (
	"errors"
	"fmt"
	"pupload/internal/models"
	"slices"
)

func ValidateFlow(flow models.Flow, nodeDefs []models.NodeDef) error {

	edgeCount := generateEdgeCount(flow)

	// Check that all node definitions exist

	// Check that all edges are valid

	// Check for cycles

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

	for _, node := range flow.Nodes {
		def := new(models.NodeDef)

		for _, d := range nodeDefs {
			if d.Name == node.DefName {
				def = &d
			}
		}

		if def == nil {
			return ErrNodeDefNotFound
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

	return true
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
