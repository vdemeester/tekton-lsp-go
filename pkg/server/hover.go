package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/hover"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// textDocumentHover handles the textDocument/hover request.
func (s *Server) textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	parsed, ok := s.cache.GetParsed(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	result := hover.Hover(parsed, parser.Position{
		Line:      params.Position.Line,
		Character: params.Position.Character,
	})

	if result == nil {
		return nil, nil
	}

	h := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: result.Content,
		},
	}

	if result.Range != nil {
		r := protocol.Range{
			Start: protocol.Position{Line: result.Range.Start.Line, Character: result.Range.Start.Character},
			End:   protocol.Position{Line: result.Range.End.Line, Character: result.Range.End.Character},
		}
		h.Range = &r
	}

	return h, nil
}
