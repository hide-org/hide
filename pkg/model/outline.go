package model

type Outline struct {
	Path  string
	Nodes []Node
}

type Node struct {
	Kind      string // The type of the element (e.g., "Class", "Function", "Variable", "Section")
	Name      string // The name of the element (e.g., "User", "calculateAge", "Introduction")
	Children  []Node
	StartLine int
	EndLine   int
}
