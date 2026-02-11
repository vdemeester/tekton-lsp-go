package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// didOpen handles the textDocument/didOpen notification
func (s *Server) didOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	text := params.TextDocument.Text

	log.Infof("Document opened: %s (%d bytes)", uri, len(text))

	// TODO: Parse document with tree-sitter
	// TODO: Store in document cache
	// TODO: Run validation and send diagnostics

	return nil
}

// didChange handles the textDocument/didChange notification
func (s *Server) didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI

	log.Infof("Document changed: %s", uri)

	// With TextDocumentSyncKindFull, there's only one change with the full content
	if len(params.ContentChanges) > 0 {
		// Get the new full content
		if textChange, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEvent); ok {
			log.Debugf("New content: %d bytes", len(textChange.Text))

			// TODO: Re-parse document with tree-sitter
			// TODO: Update document cache
			// TODO: Re-run validation and send diagnostics
		}
	}

	return nil
}

// didClose handles the textDocument/didClose notification
func (s *Server) didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI

	log.Infof("Document closed: %s", uri)

	// TODO: Remove from document cache

	return nil
}
