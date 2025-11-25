package models

type NodeDef struct {
	ID        int64
	Publisher string
	Name      string
	Image     string
	Inputs    []NodeEdgeDef
	Outputs   []NodeEdgeDef
	Flags     []NodeFlagDef
	Command   NodeCommandDef
}

type NodeFlagDef struct {
	Name        string
	Description string
	Required    bool
	Type        string
}

type NodeEdgeDef struct {
	Name        string
	Description string
	Required    bool
	Type        []string
}

type NodeCommandDef struct {
	Name        string
	Description string
	Exec        []string
}

type Node struct {
	ID      int
	DefName string
	Inputs  []NodeEdge
	Outputs []NodeEdge
	Flags   []NodeFlag
	Command string
}

type NodeEdge struct {
	Name string
	Edge string

	Store *string // optional
}

type NodeFlag struct {
	Name  string
	Value string
}
