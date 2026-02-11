package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
	"github.com/vdemeester/tekton-lsp-go/pkg/validator"
)

func TestConvertDiagnostics(t *testing.T) {
	input := []validator.Diagnostic{
		{
			Range: parser.Range{
				Start: parser.Position{Line: 2, Character: 0},
				End:   parser.Position{Line: 2, Character: 10},
			},
			Severity: validator.SeverityError,
			Source:   "tekton-lsp",
			Message:  "Required field 'metadata.name' is missing",
		},
		{
			Range: parser.Range{
				Start: parser.Position{Line: 5, Character: 2},
				End:   parser.Position{Line: 5, Character: 7},
			},
			Severity: validator.SeverityWarning,
			Source:   "tekton-lsp",
			Message:  "Unknown field 'taskz' in spec",
		},
	}

	result := convertDiagnostics(input)

	require.Len(t, result, 2)

	// Error diagnostic
	assert.Equal(t, protocol.DiagnosticSeverityError, *result[0].Severity)
	assert.Equal(t, "Required field 'metadata.name' is missing", result[0].Message)
	assert.Equal(t, uint32(2), result[0].Range.Start.Line)
	assert.Equal(t, uint32(0), result[0].Range.Start.Character)
	source0 := "tekton-lsp"
	assert.Equal(t, &source0, result[0].Source)

	// Warning diagnostic
	assert.Equal(t, protocol.DiagnosticSeverityWarning, *result[1].Severity)
	assert.Equal(t, "Unknown field 'taskz' in spec", result[1].Message)
	assert.Equal(t, uint32(5), result[1].Range.Start.Line)
}

func TestConvertDiagnostics_Empty(t *testing.T) {
	result := convertDiagnostics(nil)
	assert.Empty(t, result)
}

func TestValidateAndCollect(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	// Insert a document with an error
	s.cache.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  namespace: default
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)

	diags := s.validateDocument("file:///test.yaml")
	require.NotEmpty(t, diags, "should return diagnostics for invalid document")
	assert.Equal(t, protocol.DiagnosticSeverityError, *diags[0].Severity)
	assert.Contains(t, diags[0].Message, "metadata.name")
}

func TestValidateAndCollect_ValidDocument(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	s.cache.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: my-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)

	diags := s.validateDocument("file:///test.yaml")
	assert.Empty(t, diags, "valid document should produce no diagnostics")
}

func TestValidateAndCollect_MissingDocument(t *testing.T) {
	s := New("test-lsp", "0.1.0")

	diags := s.validateDocument("file:///nonexistent.yaml")
	assert.Empty(t, diags, "missing document should return empty diagnostics")
}
