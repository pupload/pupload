package models

type Flow struct {
	Name string

	Stores []StoreInput

	DefaultDataWell *DataWell

	DataWells []DataWell
	Nodes     []Node
}

func (f *Flow) Normalize() {
	for _, s := range f.Stores {
		s.Normalize()
	}

}
