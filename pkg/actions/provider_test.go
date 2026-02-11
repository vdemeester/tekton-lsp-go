package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
	"github.com/vdemeester/tekton-lsp-go/pkg/validator"
)

func makeDiag(msg string, line uint32, severity validator.Severity) validator.Diagnostic {
	return validator.Diagnostic{
		Range: parser.Range{
			Start: parser.Position{Line: line, Character: 0},
			End:   parser.Position{Line: line, Character: 10},
		},
		Severity: severity,
		Source:   "tekton-lsp",
		Message:  msg,
	}
}

func TestCodeActions_AddMissingField(t *testing.T) {
	diag := makeDiag("Required field 'metadata.name' is missing", 2, validator.SeverityError)
	actions := CodeActions("file:///test.yaml", []validator.Diagnostic{diag})

	require.Len(t, actions, 1)
	assert.Contains(t, actions[0].Title, "Add")
	assert.Contains(t, actions[0].Title, "name")
	assert.Equal(t, CodeActionKindQuickFix, actions[0].Kind)
	assert.NotEmpty(t, actions[0].NewText, "should provide text to insert")
}

func TestCodeActions_RemoveUnknownField(t *testing.T) {
	diag := makeDiag("Unknown field 'taskz' in spec", 5, validator.SeverityWarning)
	actions := CodeActions("file:///test.yaml", []validator.Diagnostic{diag})

	require.Len(t, actions, 1)
	assert.Contains(t, actions[0].Title, "Remove")
	assert.Contains(t, actions[0].Title, "taskz")
	assert.Equal(t, CodeActionKindQuickFix, actions[0].Kind)
}

func TestCodeActions_NoActionForUnhandled(t *testing.T) {
	diag := makeDiag("Pipeline must have at least one task", 5, validator.SeverityError)
	actions := CodeActions("file:///test.yaml", []validator.Diagnostic{diag})

	assert.Empty(t, actions, "no action for unhandled diagnostics")
}

func TestCodeActions_MultipleActions(t *testing.T) {
	diags := []validator.Diagnostic{
		makeDiag("Required field 'metadata' is missing", 0, validator.SeverityError),
		makeDiag("Unknown field 'foo' in spec", 5, validator.SeverityWarning),
	}
	actions := CodeActions("file:///test.yaml", diags)

	assert.Len(t, actions, 2, "should create action per actionable diagnostic")
}

func TestCodeActions_EmptyDiagnostics(t *testing.T) {
	actions := CodeActions("file:///test.yaml", nil)
	assert.Empty(t, actions)
}
