package models

const (
	DefaultTier     = "c-small"
	DefaultAttempts = 3
)

type NodeDef struct {
	ID          int64
	Publisher   string
	Name        string
	Image       string
	Inputs      []NodeEdgeDef
	Outputs     []NodeEdgeDef
	Flags       []NodeFlagDef
	Command     NodeCommandDef
	Tier        string
	MaxAttempts int
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
	Type        []MimeType
}

type NodeCommandDef struct {
	Name        string
	Description string
	Exec        string
}

func (nd *NodeDef) Normalize() {
	if nd.Tier == "" {
		nd.Tier = DefaultTier
	}

	if nd.MaxAttempts <= 0 {
		nd.MaxAttempts = DefaultAttempts
	}
}
