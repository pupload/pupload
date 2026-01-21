package validation

import "github.com/pupload/pupload/internal/models"

type ValidationSeverity string

const (
	ValidationError   ValidationSeverity = "ValidationError"
	ValidationWarning ValidationSeverity = "ValidationWarning"
)

type ValidationEntry struct {
	Type        ValidationSeverity
	Code        string
	Name        string
	Description string
}

type ValidationResult struct {
	Errors   []ValidationEntry
	Warnings []ValidationEntry
}

func Validate(flow models.Flow, defs []models.NodeDef) *ValidationResult {
	res := &ValidationResult{}

	// Store errors and warnings
	for _, store := range flow.Stores {
		storeInvalidType(res, store)
	}

	// Node errors and warnings
	nodeDuplicateID(res, flow.Nodes)
	for _, node := range flow.Nodes {
		nodeNoDefFound(res, node, defs)
		nodeMissingInput(res, node, defs)
		// nodeMissingOutput(res, node, defs)
		nodeInvalidTier(res, node, defs)
		nodeMissingFlag(res, node, defs)
		nodeUnknownFlag(res, node, defs)
		nodeMissingID(res, node)
	}

	// Edge errors and warnings
	edgeNoConsumers(res, flow)
	edgeNoProducers(res, flow)
	edgeTypeMismatch(res, flow, defs)

	// Well errors and warnings
	wellDuplicateEdge(res, flow.DataWells)
	wellEdgeNotFound(res, flow.DataWells, flow.Nodes)
	for _, well := range flow.DataWells {
		wellInvalidSource(res, well)
		wellStoreNotFound(res, well, flow.Stores)
		wellStaticMissingKey(res, well)
	}

	// Flow errors and warnings
	flowDetectEmpty(res, flow)
	flowDetectCycle(res, flow)

	return res
}

func (r *ValidationResult) HasError() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) HasWarnings() bool {
	return len(r.Errors)+len(r.Warnings) > 0
}

func (r *ValidationResult) AddError(entry ValidationEntry) {
	r.Errors = append(r.Errors, entry)
}

func (r *ValidationResult) AddWarning(entry ValidationEntry) {
	r.Warnings = append(r.Warnings, entry)
}
