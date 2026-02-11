package hover

import (
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// HoverResult contains the hover information for a position.
type HoverResult struct {
	Content string
	Range   *parser.Range
}

// Hover returns documentation for the node at the given position.
func Hover(doc *parser.Document, pos parser.Position) *HoverResult {
	if !strings.Contains(doc.APIVersion, "tekton.dev") {
		return nil
	}

	// Find the node at the position.
	node := doc.FindNodeAtPosition(pos)
	if node == nil {
		return nil
	}

	// Try looking up documentation by the node's key.
	if node.Key != "" {
		if content, ok := getDocumentation(node.Key); ok {
			r := node.Range
			return &HoverResult{Content: content, Range: &r}
		}
	}

	// For scalar values, try the parent key context.
	if node.IsScalar() {
		// The value itself might be interesting (e.g. "Pipeline" for kind).
		val := node.AsScalar()
		// Check if it's a known kind.
		switch val {
		case "Pipeline", "Task", "ClusterTask", "PipelineRun", "TaskRun",
			"TriggerTemplate", "TriggerBinding", "EventListener":
			content := "**" + val + "** — A Tekton " + val + " resource."
			r := node.Range
			return &HoverResult{Content: content, Range: &r}
		}
	}

	return nil
}
