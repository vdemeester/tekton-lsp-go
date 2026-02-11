package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/completion"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// textDocumentCompletion handles the textDocument/completion request.
func (s *Server) textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	return s.handleCompletion(params.TextDocument.URI, params.Position), nil
}

// handleCompletion returns completion items for a position in a document.
func (s *Server) handleCompletion(uri string, pos protocol.Position) any {
	parsed, ok := s.cache.GetParsed(uri)
	if !ok {
		return nil
	}

	items := completion.Complete(parsed, parser.Position{
		Line:      pos.Line,
		Character: pos.Character,
	})

	if len(items) == 0 {
		return nil
	}

	result := make([]protocol.CompletionItem, len(items))
	for i, item := range items {
		detail := item.Detail
		result[i] = protocol.CompletionItem{
			Label:  item.Label,
			Kind:   completionItemKind(item.Kind),
			Detail: &detail,
		}
	}
	return result
}

func completionItemKind(ft completion.FieldType) *protocol.CompletionItemKind {
	var kind protocol.CompletionItemKind
	switch ft {
	case completion.FieldTypeString:
		kind = protocol.CompletionItemKindField
	case completion.FieldTypeArray:
		kind = protocol.CompletionItemKindValue
	case completion.FieldTypeObject:
		kind = protocol.CompletionItemKindStruct
	default:
		kind = protocol.CompletionItemKindField
	}
	return &kind
}
