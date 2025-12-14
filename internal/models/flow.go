package models

type Flow struct {
	Name string

	Stores         []StoreInput
	DefaultStore   *string
	DataWells      []DataWell
	AvailableNodes []string
	Nodes          []Node
}
