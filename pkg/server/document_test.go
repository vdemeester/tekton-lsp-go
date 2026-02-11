package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestDidChange_UpdatesCacheWithFullSync(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	// Simulate didOpen
	s.cache.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  namespace: default
spec:
  steps:
    - name: build
      image: golang:1.25
`)

	// Verify initial state has error (missing metadata.name)
	diags := s.validateDocument("file:///test.yaml")
	require.NotEmpty(t, diags, "should have diagnostics before fix")

	// Simulate didChange with full content (TextDocumentContentChangeEventWhole)
	newContent := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  steps:
    - name: build
      image: golang:1.25
`
	// This is how GLSP delivers full-sync changes
	change := protocol.TextDocumentContentChangeEventWhole{
		Text: newContent,
	}
	s.handleContentChange("file:///test.yaml", 2, []any{change})

	// After fix, diagnostics should be empty
	diags = s.validateDocument("file:///test.yaml")
	assert.Empty(t, diags, "diagnostics should clear after fixing the document")

	// Verify cache has updated content
	entry, ok := s.cache.Get("file:///test.yaml")
	require.True(t, ok)
	assert.Contains(t, entry.Content, "name: my-task")
	assert.Equal(t, int32(2), entry.Version)
}

func TestDidChange_HandlesBothChangeTypes(t *testing.T) {
	s := New("test-lsp", "0.1.0")
	s.cache.Insert("file:///test.yaml", "yaml", 1, "old content")

	// TextDocumentContentChangeEvent (with range — partial sync)
	changePartial := protocol.TextDocumentContentChangeEvent{
		Text: "new content via partial",
	}
	s.handleContentChange("file:///test.yaml", 2, []any{changePartial})

	entry, _ := s.cache.Get("file:///test.yaml")
	assert.Equal(t, "new content via partial", entry.Content)

	// TextDocumentContentChangeEventWhole (no range — full sync)
	changeFull := protocol.TextDocumentContentChangeEventWhole{
		Text: "new content via full",
	}
	s.handleContentChange("file:///test.yaml", 3, []any{changeFull})

	entry, _ = s.cache.Get("file:///test.yaml")
	assert.Equal(t, "new content via full", entry.Content)
}
