package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/actions"
	"github.com/vdemeester/tekton-lsp-go/pkg/validator"
)

// textDocumentCodeAction handles the textDocument/codeAction request.
func (s *Server) textDocumentCodeAction(context *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	parsed, ok := s.cache.GetParsed(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	diags := validator.Validate(parsed)
	codeActions := actions.CodeActions(params.TextDocument.URI, diags)

	if len(codeActions) == 0 {
		return nil, nil
	}

	result := make([]protocol.CodeAction, len(codeActions))
	for i, a := range codeActions {
		edit := map[string][]protocol.TextEdit{
			a.URI: {
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: a.Range.Start.Line, Character: a.Range.Start.Character},
						End:   protocol.Position{Line: a.Range.End.Line, Character: a.Range.End.Character},
					},
					NewText: a.NewText,
				},
			},
		}
		kind := protocol.CodeActionKindQuickFix
		result[i] = protocol.CodeAction{
			Title: a.Title,
			Kind:  &kind,
			Edit: &protocol.WorkspaceEdit{
				Changes: edit,
			},
		}
	}
	return result, nil
}
