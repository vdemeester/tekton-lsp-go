package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestServer_Completion_PipelineSpec(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	s.cache.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  
`)

	result := s.handleCompletion("file:///test.yaml", protocol.Position{Line: 5, Character: 2})
	require.NotNil(t, result)

	items, ok := result.([]protocol.CompletionItem)
	require.True(t, ok, "result should be []CompletionItem")
	require.NotEmpty(t, items)

	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label
	}
	assert.Contains(t, labels, "tasks")
	assert.Contains(t, labels, "params")
}

func TestServer_Completion_MissingDocument(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	result := s.handleCompletion("file:///nonexistent.yaml", protocol.Position{Line: 0, Character: 0})
	assert.Nil(t, result, "missing document should return nil")
}
