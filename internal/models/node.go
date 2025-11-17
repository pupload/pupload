package models

type NodeDef struct {
	ID        int64
	Publisher string
	Name      string
	Inputs    []NodeEdgeDef
	Outputs   []NodeEdgeDef
	Flags     []NodeFlagDef
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

type Node struct {
	ID      int
	DefName string
	Inputs  []NodeEdge
	Outputs []NodeEdge
}

type NodeEdge struct {
	ID       int
	Name     string
	Required bool
	Value    interface{}
}

type NodeExecutePayload struct {
	NodeDef NodeDef
	Node    Node
}
