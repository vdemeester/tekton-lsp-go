package completion

import (
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// contextKind identifies where in the YAML structure the cursor is.
type contextKind int

const (
	contextUnknown contextKind = iota
	contextMetadata
	contextPipelineSpec
	contextTaskSpec
	contextPipelineTask
	contextStep
)

// Complete returns completion items for the given position in the document.
func Complete(doc *parser.Document, pos parser.Position) []CompletionItem {
	if !isTekton(doc) {
		return nil
	}

	ctx := determineContext(doc, pos)
	fields := fieldsForContext(ctx)

	items := make([]CompletionItem, len(fields))
	for i, f := range fields {
		items[i] = CompletionItem{
			Label:  f.Name,
			Detail: f.Description,
			Kind:   f.Type,
		}
	}
	return items
}

func isTekton(doc *parser.Document) bool {
	return strings.Contains(doc.APIVersion, "tekton.dev")
}

func determineContext(doc *parser.Document, pos parser.Position) contextKind {
	// First try exact range matching.
	ctx := walkForContext(doc.Root, pos, doc.Kind)
	if ctx != contextUnknown {
		return ctx
	}

	// Fallback: find the nearest key above the cursor position in the root mapping.
	// This handles cases where the cursor is on an empty/blank line under a key
	// (tree-sitter ranges don't extend to cover blank lines).
	return findContextByNearestKey(doc.Root, pos, doc.Kind)
}

// findContextByNearestKey looks at the root mapping's children and finds the key
// whose line is closest above the cursor. This handles empty block values.
func findContextByNearestKey(root *parser.Node, pos parser.Position, kind string) contextKind {
	if !root.IsMapping() {
		return contextUnknown
	}

	var bestKey string
	var bestLine uint32

	for key, child := range root.MappingChildren {
		startLine := child.Range.Start.Line
		if startLine <= pos.Line && startLine >= bestLine {
			bestKey = key
			bestLine = startLine
		}
	}

	switch bestKey {
	case "metadata":
		return contextMetadata
	case "spec":
		// Check if cursor is indented enough to be inside spec children.
		spec := root.Get("spec")
		if spec != nil && spec.IsMapping() {
			sub := findContextByNearestKey(spec, pos, kind)
			if sub != contextUnknown {
				return sub
			}
		}
		switch kind {
		case "Pipeline":
			return contextPipelineSpec
		case "Task", "ClusterTask":
			return contextTaskSpec
		}
	case "tasks", "finally":
		return contextPipelineTask
	case "steps":
		return contextStep
	}

	return contextUnknown
}

func walkForContext(node *parser.Node, pos parser.Position, kind string) contextKind {
	if node == nil || !posInRange(pos, node.Range) {
		return contextUnknown
	}

	// Check named mapping children for context clues.
	if node.IsMapping() {
		for key, child := range node.MappingChildren {
			if !posInRange(pos, child.Range) {
				continue
			}

			switch key {
			case "metadata":
				return contextMetadata
			case "spec":
				// Recurse into spec to see if we're in a sub-field.
				sub := walkSpecChildren(child, pos)
				if sub != contextUnknown {
					return sub
				}
				// Top-level spec context.
				switch kind {
				case "Pipeline":
					return contextPipelineSpec
				case "Task", "ClusterTask":
					return contextTaskSpec
				}
			case "tasks", "finally":
				// Inside a tasks/finally array item.
				return contextPipelineTask
			case "steps":
				return contextStep
			}

			// Recurse deeper.
			deeper := walkForContext(child, pos, kind)
			if deeper != contextUnknown {
				return deeper
			}
		}
	}

	if node.IsSequence() {
		for _, child := range node.SequenceChildren {
			deeper := walkForContext(child, pos, kind)
			if deeper != contextUnknown {
				return deeper
			}
		}
	}

	return contextUnknown
}

// walkSpecChildren checks spec's immediate children for tasks/steps arrays.
func walkSpecChildren(spec *parser.Node, pos parser.Position) contextKind {
	if !spec.IsMapping() {
		return contextUnknown
	}

	for key, child := range spec.MappingChildren {
		if !posInRange(pos, child.Range) {
			continue
		}
		switch key {
		case "tasks", "finally":
			return contextPipelineTask
		case "steps":
			return contextStep
		}
	}
	return contextUnknown
}

func fieldsForContext(ctx contextKind) []FieldSchema {
	switch ctx {
	case contextMetadata:
		return metadataFields
	case contextPipelineSpec:
		return pipelineSpecFields
	case contextTaskSpec:
		return taskSpecFields
	case contextPipelineTask:
		return pipelineTaskFields
	case contextStep:
		return stepFields
	default:
		return nil
	}
}

func posInRange(pos parser.Position, r parser.Range) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	return true
}
