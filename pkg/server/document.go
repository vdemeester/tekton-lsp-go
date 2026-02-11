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

	s.publishDiagnostics(context, uri)

	return nil
}

// didChange handles the textDocument/didChange notification
func (s *Server) didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	version := params.TextDocument.Version

	log.Infof("Document changed: %s (v%d)", uri, version)

	s.handleContentChange(uri, version, params.ContentChanges)
	s.publishDiagnostics(context, uri)

	return nil
}

// handleContentChange extracts the new content from change events and updates the cache.
// GLSP delivers full-sync changes as TextDocumentContentChangeEventWhole
// and partial-sync changes as TextDocumentContentChangeEvent.
func (s *Server) handleContentChange(uri string, version int32, changes []any) {
	if len(changes) == 0 {
		return
	}

	switch change := changes[0].(type) {
	case protocol.TextDocumentContentChangeEventWhole:
		s.cache.Update(uri, version, change.Text)
	case protocol.TextDocumentContentChangeEvent:
		s.cache.Update(uri, version, change.Text)
	}
}

// didClose handles the textDocument/didClose notification
func (s *Server) didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI

	log.Infof("Document closed: %s", uri)

	s.cache.Remove(uri)

	return nil
}
