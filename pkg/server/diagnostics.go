package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/validator"
)

// validateDocument runs validation on all documents in a file and returns LSP diagnostics.
func (s *Server) validateDocument(uri string) []protocol.Diagnostic {
	docs, ok := s.cache.GetAllParsed(uri)
	if !ok {
		return []protocol.Diagnostic{}
	}

	var allDiags []validator.Diagnostic
	for _, doc := range docs {
		allDiags = append(allDiags, validator.Validate(doc)...)
	}
	return convertDiagnostics(allDiags)
}

// publishDiagnostics sends diagnostics to the LSP client.
func (s *Server) publishDiagnostics(context *glsp.Context, uri string) {
	diags := s.validateDocument(uri)

	go context.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}

// convertDiagnostics converts our validator diagnostics to LSP protocol diagnostics.
func convertDiagnostics(diags []validator.Diagnostic) []protocol.Diagnostic {
	if len(diags) == 0 {
		return []protocol.Diagnostic{}
	}

	result := make([]protocol.Diagnostic, len(diags))
	for i, d := range diags {
		severity := convertSeverity(d.Severity)
		source := d.Source
		result[i] = protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      d.Range.Start.Line,
					Character: d.Range.Start.Character,
				},
				End: protocol.Position{
					Line:      d.Range.End.Line,
					Character: d.Range.End.Character,
				},
			},
			Severity: &severity,
			Source:   &source,
			Message:  d.Message,
		}
	}
	return result
}

func convertSeverity(s validator.Severity) protocol.DiagnosticSeverity {
	switch s {
	case validator.SeverityError:
		return protocol.DiagnosticSeverityError
	case validator.SeverityWarning:
		return protocol.DiagnosticSeverityWarning
	case validator.SeverityInfo:
		return protocol.DiagnosticSeverityInformation
	case validator.SeverityHint:
		return protocol.DiagnosticSeverityHint
	default:
		return protocol.DiagnosticSeverityError
	}
}
