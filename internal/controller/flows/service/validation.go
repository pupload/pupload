package service

import (
	"fmt"
	"pupload/internal/models"
)

func (f *FlowService) ValidateFlow(flow *models.Flow) error {

	// Check that all node definitions exist

	// Check that all edges are valid

	// Check for cycles

	// Check that all stores are valid

	// Check that all data wells are valid
	for _, well := range flow.DataWells {
		if !(well.Type == "static" || well.Type == "dynamic") {
			return fmt.Errorf("Data well %s has invalid type %s", well.Edge, well.Type)
		}

		if well.Type == "static" && well.Key == nil {
			return fmt.Errorf("Data well %s is static but has no key", well.Edge)
		}

		if well.Type == "dynamic" && well.Key != nil && validateKey(*well.Key) {
			return fmt.Errorf("Data well %s key is invalid", well.Edge)
		}
	}

	// Check if DefaultStore is set
	if flow.DefaultStore == nil {
		fmt.Println(flow)

		if len(flow.Stores) == 0 {
			return fmt.Errorf("no stores are defined")
		}

		storeName := flow.Stores[0]
		flow.DefaultStore = &storeName.Name
	}

	return nil
}

func validateKey(key string) bool {
	return len(key) > 0
}
