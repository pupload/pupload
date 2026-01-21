package validation

import (
	"fmt"

	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/resources"
)

func getNodeDef(node models.Node, defs []models.NodeDef) *models.NodeDef {
	var def *models.NodeDef

	for _, d := range defs {
		if fmt.Sprintf("%s/%s", d.Publisher, d.Name) == node.Uses {
			def = &d
		}
	}

	return def
}

func nodeNoDefFound(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeDefNotFound,
			"NodeDefNotFound",
			fmt.Sprintf("Definition %s not found (node %s)", node.Uses, node.ID),
		})
	}
}

func nodeMissingInput(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		return
	}

	for _, inDef := range def.Inputs {
		found := false
		for _, inNode := range node.Inputs {
			if !inDef.Required || inNode.Name == inDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingInput,
			"NodeMissingInput",
			fmt.Sprintf("Node %s missing required input %s", node.ID, inDef.Name),
		})

	}
}

// should be a runtime error, the required flag should mean that the worker doesn't generate an output
func nodeMissingOutput(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		return
	}

	for _, outDef := range def.Outputs {
		found := false
		for _, outNode := range node.Outputs {
			if !outDef.Required || outNode.Name == outDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingInput,
			"NodeMissingOutput",
			fmt.Sprintf("Node %s missing required output %s", node.ID, outDef.Name),
		})
	}
}

func nodeInvalidTier(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		return
	}

	_, tierValid := resources.StandardTierMap[def.Tier]
	if !tierValid {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingInput,
			"NodeInvalidTier",
			fmt.Sprintf("Node %s uses invalid tier %s", node.ID, def.Tier),
		})
	}
}

func nodeMissingFlag(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		return
	}

	for _, flagDef := range def.Flags {
		found := false
		for _, flagNode := range node.Flags {
			if !flagDef.Required || flagNode.Name == flagDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingFlag,
			"NodeMissingFlag",
			fmt.Sprintf("Node %s missing required flag %s", node.ID, flagDef.Name),
		})
	}
}

func nodeUnknownFlag(r *ValidationResult, node models.Node, defs []models.NodeDef) {
	def := getNodeDef(node, defs)
	if def == nil {
		return
	}

	for _, flagNode := range node.Flags {
		found := false
		for _, flagDef := range def.Flags {
			if flagNode.Name == flagDef.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingFlag,
			"NodeUnknownFlag",
			fmt.Sprintf("Node %s has unknown flag %s", node.ID, flagNode.Name),
		})
	}
}

func nodeMissingID(r *ValidationResult, node models.Node) {
	if node.ID == "" {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrNodeMissingID,
			"NodeMissingID",
			"Node missing ID",
		})
	}
}

func nodeDuplicateID(r *ValidationResult, nodes []models.Node) {
	idCount := make(map[string]int)
	for _, node := range nodes {
		idCount[node.ID]++
	}

	for id, val := range idCount {
		if val > 1 {
			r.AddError(ValidationEntry{
				ValidationError,
				ErrNodeDuplicateID,
				"NodeDuplicateID",
				fmt.Sprintf("Node ID %s is used more than once", id),
			})
		}
	}

}
