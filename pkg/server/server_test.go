package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	assert.NotNil(t, s)
	assert.Equal(t, "test-lsp", s.name)
	assert.Equal(t, "0.1.0", s.version)
	assert.NotNil(t, s.glsp)
}

func TestServer_HasDocumentCache(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	require.NotNil(t, s.cache, "server should have a document cache")
}
