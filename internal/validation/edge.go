package validation

import (
	"fmt"

	mimetypes "github.com/pupload/pupload/internal/mimetype"
	"github.com/pupload/pupload/internal/models"
)

func edgeNoProducers(r *ValidationResult, flow models.Flow) {
	hasProducer := make(map[string]bool)
	for _, node := range flow.Nodes {
		for _, input := range node.Inputs {
			hasProducer[input.Edge] = false
		}
	}

	for _, well := range flow.DataWells {
		if well.Source != nil {
			hasProducer[well.Edge] = true
		}
	}

	for _, node := range flow.Nodes {
		for _, output := range node.Outputs {
			hasProducer[output.Edge] = true
		}
	}

	for edgeName, consumed := range hasProducer {
		if !consumed {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrEdgeNoProducer,
				"EdgeNoProducer",
				fmt.Sprintf("Edge %s not bound to any output", edgeName),
			})
		}
	}

}

func edgeNoConsumers(r *ValidationResult, flow models.Flow) {
	hasConsumer := make(map[string]bool)
	for _, node := range flow.Nodes {
		for _, output := range node.Outputs {
			hasConsumer[output.Edge] = false
		}
	}

	for _, well := range flow.DataWells {
		if well.Source == nil {
			hasConsumer[well.Edge] = true
		}
	}

	for _, node := range flow.Nodes {
		for _, input := range node.Inputs {
			hasConsumer[input.Edge] = true
		}
	}

	for edgeName, consumed := range hasConsumer {
		if !consumed {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrEdgeNoConsumer,
				"EdgeNoConsumer",
				fmt.Sprintf("Edge %s not bound to any input", edgeName),
				// "Flow has a cycle, which would cause infinite execution",
			})
		}
	}
}

func edgeTypeMismatch(r *ValidationResult, flow models.Flow, nodeDefs []models.NodeDef) {

	// edgeTypeMap := make(map[string]string)

	type EdgeNodeKey [2]string

	inputSet := make(map[EdgeNodeKey]mimetypes.MimeSet)
	outputSet := make(map[string]mimetypes.MimeSet)

	// Track edges that come from datawells (which are type-agnostic)
	datawellEdges := make(map[string]bool)
	for _, well := range flow.DataWells {
		if well.Source != nil {
			datawellEdges[well.Edge] = true
		}
	}

	for _, node := range flow.Nodes {

		var def *models.NodeDef

		for _, d := range nodeDefs {
			if d.Name == node.Uses {
				def = &d
			}
		}

		if def == nil {
			continue
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
				return
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
				return
			}

			outputSet[outEdge.Edge] = *set
		}
	}

	for _, node := range flow.Nodes {
		for _, in := range node.Inputs {
			// Skip type check for edges from datawells (they can be any type)
			if datawellEdges[in.Edge] {
				continue
			}

			inputTypes := inputSet[EdgeNodeKey{in.Edge, node.ID}]
			outputTypes := outputSet[in.Edge]

			intersection := inputTypes.Intersection(outputTypes)
			if intersection.IsEmpty() {
				r.AddError(ValidationEntry{
					ValidationError,
					ErrEdgeTypeMismatch,
					"EdgeTypeMismatch",
					fmt.Sprintf("Type mismatch on edge %s", in.Edge),
				})
			}
		}
	}

}
