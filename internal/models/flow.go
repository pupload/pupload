package models

type Flow struct {
	Stores         []StoreInput
	DefaultStore   *string
	DataWells      []DataWell
	AvailableNodes []string
	Nodes          []Node
}
