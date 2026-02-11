package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/definition"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// textDocumentDefinition handles the textDocument/definition request.
func (s *Server) textDocumentDefinition(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	parsed, ok := s.cache.GetParsed(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	loc := definition.GotoDefinition(parsed, parser.Position{
		Line:      params.Position.Line,
		Character: params.Position.Character,
	}, s.cache)

	if loc == nil {
		return nil, nil
	}

	return protocol.Location{
		URI: loc.URI,
		Range: protocol.Range{
			Start: protocol.Position{Line: loc.Range.Start.Line, Character: loc.Range.Start.Character},
			End:   protocol.Position{Line: loc.Range.End.Line, Character: loc.Range.End.Character},
		},
	}, nil
}
