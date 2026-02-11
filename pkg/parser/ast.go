package parser

// Position represents a position in a text document (0-indexed).
type Position struct {
	Line      uint32
	Character uint32
}

// Range represents a range in a text document.
type Range struct {
	Start Position
	End   Position
}

// NodeKind represents the type of a YAML node.
type NodeKind int

const (
	// NodeKindScalar is a scalar value (string, number, boolean, null).
	NodeKindScalar NodeKind = iota
	// NodeKindMapping is a mapping of keys to nodes.
	NodeKindMapping
	// NodeKindSequence is a sequence of nodes.
	NodeKindSequence
	// NodeKindNull is a null value.
	NodeKindNull
)

// Node represents a node in the YAML AST with position information.
type Node struct {
	// Key is the key for this node (if it's a map entry).
	Key string
	// Kind is the type of this node.
	Kind NodeKind
	// ScalarValue holds the value for scalar nodes.
	ScalarValue string
	// MappingChildren holds key→node pairs for mapping nodes.
	MappingChildren map[string]*Node
	// SequenceChildren holds ordered items for sequence nodes.
	SequenceChildren []*Node
	// Range is the range in the document where this node appears.
	Range Range
}

// Get returns a child node by key (for mappings). Returns nil if not found.
func (n *Node) Get(key string) *Node {
	if n.Kind != NodeKindMapping {
		return nil
	}
	return n.MappingChildren[key]
}

// AsScalar returns the scalar value as a string. Returns "" if not a scalar.
func (n *Node) AsScalar() string {
	if n.Kind != NodeKindScalar {
		return ""
	}
	return n.ScalarValue
}

// AsSequence returns the sequence items. Returns nil if not a sequence.
func (n *Node) AsSequence() []*Node {
	if n.Kind != NodeKindSequence {
		return nil
	}
	return n.SequenceChildren
}

// IsMapping returns true if this node is a mapping.
func (n *Node) IsMapping() bool {
	return n.Kind == NodeKindMapping
}

// IsSequence returns true if this node is a sequence.
func (n *Node) IsSequence() bool {
	return n.Kind == NodeKindSequence
}

// IsScalar returns true if this node is a scalar.
func (n *Node) IsScalar() bool {
	return n.Kind == NodeKindScalar
}

// Document represents a parsed YAML document.
type Document struct {
	// Filename is the URI or path of the document.
	Filename string
	// Root is the root node of the document.
	Root *Node
	// APIVersion extracted for quick access.
	APIVersion string
	// Kind extracted for quick access.
	Kind string
}

// FindNodeAtPosition returns the most specific node at the given position.
func (d *Document) FindNodeAtPosition(pos Position) *Node {
	return findNodeAtPosition(d.Root, pos)
}

func findNodeAtPosition(node *Node, pos Position) *Node {
	if node == nil {
		return nil
	}
	if !positionInRange(pos, node.Range) {
		return nil
	}

	// Depth-first: check children for a more specific match.
	switch node.Kind {
	case NodeKindMapping:
		for _, child := range node.MappingChildren {
			if found := findNodeAtPosition(child, pos); found != nil {
				return found
			}
		}
	case NodeKindSequence:
		for _, child := range node.SequenceChildren {
			if found := findNodeAtPosition(child, pos); found != nil {
				return found
			}
		}
	}

	return node
}

func positionInRange(pos Position, r Range) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	return true
}
