package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/formatting"
)

// textDocumentFormatting handles the textDocument/formatting request.
func (s *Server) textDocumentFormatting(context *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	entry, ok := s.cache.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	indent := 2
	if tabSize, ok := params.Options["tabSize"]; ok {
		if ts, ok := tabSize.(float64); ok {
			indent = int(ts)
		}
	}
	opts := formatting.Options{IndentSize: indent}

	formatted, err := formatting.Format(entry.Content, opts)
	if err != nil {
		return nil, nil // Swallow formatting errors silently
	}

	if formatted == entry.Content {
		return nil, nil // No changes needed
	}

	// Replace the entire document.
	lines := countLines(entry.Content)
	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: uint32(lines + 1), Character: 0},
			},
			NewText: formatted,
		},
	}, nil
}

func countLines(s string) int {
	n := 0
	for _, c := range s {
		if c == '\n' {
			n++
		}
	}
	return n
}
