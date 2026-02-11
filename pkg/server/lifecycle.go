package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// initialize handles the initialize request from the client
func (s *Server) initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	log.Info("Initializing Tekton LSP server")

	// Log client info
	if params.ClientInfo != nil {
		log.Infof("Client: %s %s", params.ClientInfo.Name, *params.ClientInfo.Version)
	}

	// Create server capabilities
	capabilities := s.handler.CreateServerCapabilities()

	// Configure text document sync
	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull

	// Completion
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{":", "-", " "},
	}

	// Hover
	capabilities.HoverProvider = true

	// Document Symbols
	capabilities.DocumentSymbolProvider = true
	// capabilities.HoverProvider = true
	// capabilities.DefinitionProvider = true
	// capabilities.DocumentSymbolProvider = true
	// capabilities.DocumentFormattingProvider = true
	// capabilities.CodeActionProvider = true

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    s.name,
			Version: &s.version,
		},
	}, nil
}

// initialized handles the initialized notification from the client
func (s *Server) initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	log.Info("Server initialized")
	return nil
}

// shutdown handles the shutdown request from the client
func (s *Server) shutdown(context *glsp.Context) error {
	log.Info("Shutting down Tekton LSP server")
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

// setTrace handles the setTrace notification from the client
func (s *Server) setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
