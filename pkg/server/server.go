package server

import (
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	"github.com/vdemeester/tekton-lsp-go/pkg/cache"
)

var log = commonlog.GetLogger("tekton-lsp")

// Server represents the Tekton LSP server
type Server struct {
	name    string
	version string
	glsp    *server.Server
	handler protocol.Handler
	cache   *cache.Cache
}

// New creates a new Tekton LSP server
func New(name, version string) *Server {
	s := &Server{
		name:    name,
		version: version,
		cache:   cache.New(),
	}

	// Initialize handler with lifecycle methods
	s.handler = protocol.Handler{
		Initialize:            s.initialize,
		Initialized:           s.initialized,
		Shutdown:              s.shutdown,
		SetTrace:              s.setTrace,
		TextDocumentDidOpen:   s.didOpen,
		TextDocumentDidChange: s.didChange,
		TextDocumentDidClose:  s.didClose,
	}

	// Create GLSP server
	s.glsp = server.NewServer(&s.handler, name, false)

	return s
}

// RunStdio runs the server using stdio transport
func (s *Server) RunStdio() error {
	log.Info("Starting Tekton LSP server (stdio)")
	return s.glsp.RunStdio()
}

// RunTCP runs the server using TCP transport
func (s *Server) RunTCP(address string) error {
	log.Infof("Starting Tekton LSP server (TCP: %s)", address)
	return s.glsp.RunTCP(address)
}
