package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// didOpen handles the textDocument/didOpen notification
func (s *Server) didOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	text := params.TextDocument.Text
	langID := params.TextDocument.LanguageID
	version := params.TextDocument.Version

	log.Infof("Document opened: %s (%d bytes)", uri, len(text))

	s.cache.Insert(uri, langID, version, text)

	// TODO: Run validation and send diagnostics

	return nil
}

// didChange handles the textDocument/didChange notification
func (s *Server) didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	version := params.TextDocument.Version

	log.Infof("Document changed: %s (v%d)", uri, version)

	// With TextDocumentSyncKindFull, there's one change with full content
	if len(params.ContentChanges) > 0 {
		if textChange, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEvent); ok {
			s.cache.Update(uri, version, textChange.Text)
		}
	}

	// TODO: Re-run validation and send diagnostics

	return nil
}

// didClose handles the textDocument/didClose notification
func (s *Server) didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI

	log.Infof("Document closed: %s", uri)

	s.cache.Remove(uri)

	return nil
}
