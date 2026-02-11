package definition

import (
	"github.com/vdemeester/tekton-lsp-go/pkg/cache"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// Location represents a target definition location.
type Location struct {
	URI   string
	Range parser.Range
}

// GotoDefinition resolves a taskRef/pipelineRef at the given position to its definition.
func GotoDefinition(doc *parser.Document, pos parser.Position, c *cache.Cache) *Location {
	// Find what reference we're on.
	ref := findReference(doc.Root, pos)
	if ref == nil {
		return nil
	}

	// Search all cached documents for a matching resource.
	for _, parsed := range c.AllParsed() {
		if parsed.Kind == ref.kind {
			if metadata := parsed.Root.Get("metadata"); metadata != nil {
				if name := metadata.Get("name"); name != nil && name.AsScalar() == ref.name {
					return &Location{
						URI:   parsed.Filename,
						Range: parsed.Root.Range,
					}
				}
			}
		}
	}

	return nil
}

type reference struct {
	kind string
	name string
}

// findReference walks the AST looking for a taskRef/pipelineRef at the given position.
func findReference(node *parser.Node, pos parser.Position) *reference {
	if node == nil || !posInRange(pos, node.Range) {
		return nil
	}

	if node.IsMapping() {
		for key, child := range node.MappingChildren {
			if !posInRange(pos, child.Range) {
				continue
			}

			switch key {
			case "taskRef":
				if nameNode := child.Get("name"); nameNode != nil {
					kind := "Task"
					if k := child.Get("kind"); k != nil && k.AsScalar() != "" {
						kind = k.AsScalar()
					}
					return &reference{kind: kind, name: nameNode.AsScalar()}
				}
			case "pipelineRef":
				if nameNode := child.Get("name"); nameNode != nil {
					return &reference{kind: "Pipeline", name: nameNode.AsScalar()}
				}
			}

			// Recurse deeper.
			if ref := findReference(child, pos); ref != nil {
				return ref
			}
		}
	}

	if node.IsSequence() {
		for _, child := range node.SequenceChildren {
			if ref := findReference(child, pos); ref != nil {
				return ref
			}
		}
	}

	return nil
}

func posInRange(pos parser.Position, r parser.Range) bool {
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
