package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/symbols"
)

// textDocumentDocumentSymbol handles the textDocument/documentSymbol request.
func (s *Server) textDocumentDocumentSymbol(context *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	parsed, ok := s.cache.GetParsed(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	syms := symbols.DocumentSymbols(parsed)
	if len(syms) == 0 {
		return nil, nil
	}

	return convertSymbols(syms), nil
}

func convertSymbols(syms []symbols.Symbol) []protocol.DocumentSymbol {
	result := make([]protocol.DocumentSymbol, len(syms))
	for i, s := range syms {
		r := protocol.Range{
			Start: protocol.Position{Line: s.Range.Start.Line, Character: s.Range.Start.Character},
			End:   protocol.Position{Line: s.Range.End.Line, Character: s.Range.End.Character},
		}
		result[i] = protocol.DocumentSymbol{
			Name:           s.Name,
			Kind:           convertSymbolKind(s.Kind),
			Range:          r,
			SelectionRange: r,
		}
		if len(s.Children) > 0 {
			result[i].Children = convertSymbols(s.Children)
		}
	}
	return result
}

func convertSymbolKind(kind symbols.SymbolKind) protocol.SymbolKind {
	switch kind {
	case symbols.SymbolKindObject:
		return protocol.SymbolKindObject
	case symbols.SymbolKindArray:
		return protocol.SymbolKindArray
	case symbols.SymbolKindProperty:
		return protocol.SymbolKindProperty
	default:
		return protocol.SymbolKindVariable
	}
}
