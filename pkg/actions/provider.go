package actions

import (
	"fmt"
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
	"github.com/vdemeester/tekton-lsp-go/pkg/validator"
)

// CodeActionKind identifies the type of code action.
type CodeActionKind string

const (
	CodeActionKindQuickFix CodeActionKind = "quickfix"
)

// CodeAction represents a suggested fix for a diagnostic.
type CodeAction struct {
	Title   string
	Kind    CodeActionKind
	URI     string
	Range   parser.Range
	NewText string
	Diag    validator.Diagnostic
}

// CodeActions returns quick fixes for the given diagnostics.
func CodeActions(uri string, diags []validator.Diagnostic) []CodeAction {
	var result []CodeAction

	for _, d := range diags {
		if action := actionForDiagnostic(uri, d); action != nil {
			result = append(result, *action)
		}
	}

	return result
}

func actionForDiagnostic(uri string, diag validator.Diagnostic) *CodeAction {
	msg := diag.Message

	if strings.Contains(msg, "Required field") || strings.Contains(msg, "missing") {
		return addFieldAction(uri, diag)
	}

	if strings.Contains(msg, "Unknown field") {
		return removeFieldAction(uri, diag)
	}

	return nil
}

func addFieldAction(uri string, diag validator.Diagnostic) *CodeAction {
	field := extractQuotedName(diag.Message)
	if field == "" {
		return nil
	}

	// Strip any path prefix (e.g., "metadata.name" → "name").
	parts := strings.Split(field, ".")
	shortName := parts[len(parts)-1]

	return &CodeAction{
		Title: fmt.Sprintf("Add missing field '%s'", shortName),
		Kind:  CodeActionKindQuickFix,
		URI:   uri,
		Range: parser.Range{
			Start: parser.Position{Line: diag.Range.End.Line + 1, Character: 0},
			End:   parser.Position{Line: diag.Range.End.Line + 1, Character: 0},
		},
		NewText: fieldTemplate(shortName),
		Diag:    diag,
	}
}

func removeFieldAction(uri string, diag validator.Diagnostic) *CodeAction {
	field := extractQuotedName(diag.Message)
	if field == "" {
		return nil
	}

	return &CodeAction{
		Title: fmt.Sprintf("Remove unknown field '%s'", field),
		Kind:  CodeActionKindQuickFix,
		URI:   uri,
		Range: parser.Range{
			Start: parser.Position{Line: diag.Range.Start.Line, Character: 0},
			End:   parser.Position{Line: diag.Range.Start.Line + 1, Character: 0},
		},
		NewText: "",
		Diag:    diag,
	}
}

func extractQuotedName(msg string) string {
	start := strings.Index(msg, "'")
	if start < 0 {
		return ""
	}
	rest := msg[start+1:]
	end := strings.Index(rest, "'")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func fieldTemplate(name string) string {
	switch name {
	case "metadata":
		return "metadata:\n  name: \n"
	case "name":
		return "  name: \n"
	case "spec":
		return "spec:\n  steps:\n    - name: step-1\n      image: alpine\n"
	case "steps":
		return "  steps:\n    - name: step-1\n      image: alpine\n"
	case "tasks":
		return "  tasks:\n    - name: task-1\n      taskRef:\n        name: \n"
	default:
		return fmt.Sprintf("  %s: \n", name)
	}
}
