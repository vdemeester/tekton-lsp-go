package symbols

import (
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// SymbolKind identifies the type of symbol.
type SymbolKind int

const (
	SymbolKindObject   SymbolKind = 1
	SymbolKindArray    SymbolKind = 2
	SymbolKindProperty SymbolKind = 3
)

// Symbol represents a document symbol with optional children.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	Range    parser.Range
	Children []Symbol
}

// DocumentSymbols extracts an outline of symbols from a Tekton document.
func DocumentSymbols(doc *parser.Document) []Symbol {
	if !strings.Contains(doc.APIVersion, "tekton.dev") {
		return nil
	}

	// Build root symbol from resource name and kind.
	name := doc.Kind
	if metadata := doc.Root.Get("metadata"); metadata != nil {
		if n := metadata.Get("name"); n != nil {
			name = n.AsScalar()
		}
	}

	root := Symbol{
		Name:  name,
		Kind:  SymbolKindObject,
		Range: doc.Root.Range,
	}

	// Add spec children.
	if spec := doc.Root.Get("spec"); spec != nil && spec.IsMapping() {
		for key, child := range spec.MappingChildren {
			sym := Symbol{
				Name:  key,
				Kind:  symbolKindFor(child),
				Range: child.Range,
			}

			// For arrays like tasks/steps/params, list named items.
			if child.IsSequence() {
				for _, item := range child.AsSequence() {
					if n := item.Get("name"); n != nil {
						sym.Children = append(sym.Children, Symbol{
							Name:  n.AsScalar(),
							Kind:  SymbolKindProperty,
							Range: item.Range,
						})
					}
				}
			}

			root.Children = append(root.Children, sym)
		}
	}

	return []Symbol{root}
}

func symbolKindFor(node *parser.Node) SymbolKind {
	if node.IsSequence() {
		return SymbolKindArray
	}
	if node.IsMapping() {
		return SymbolKindObject
	}
	return SymbolKindProperty
}
