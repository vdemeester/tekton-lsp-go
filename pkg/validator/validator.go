package validator

import (
	"fmt"
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// Severity represents the severity of a diagnostic.
type Severity int

const (
	SeverityError   Severity = 1
	SeverityWarning Severity = 2
	SeverityInfo    Severity = 3
	SeverityHint    Severity = 4
)

// Diagnostic represents a validation issue at a specific location.
type Diagnostic struct {
	Range    parser.Range
	Severity Severity
	Source   string
	Message  string
}

// tektonAPIVersions are the known Tekton API versions.
var tektonAPIVersions = []string{
	"tekton.dev/v1",
	"tekton.dev/v1beta1",
	"tekton.dev/v1alpha1",
	"triggers.tekton.dev/v1beta1",
	"triggers.tekton.dev/v1alpha1",
}

// knownPipelineSpecFields are the valid fields in a Pipeline spec.
var knownPipelineSpecFields = map[string]bool{
	"tasks":       true,
	"params":      true,
	"workspaces":  true,
	"results":     true,
	"finally":     true,
	"description": true,
}

// knownTaskSpecFields are the valid fields in a Task spec.
var knownTaskSpecFields = map[string]bool{
	"steps":        true,
	"params":       true,
	"workspaces":   true,
	"results":      true,
	"description":  true,
	"volumes":      true,
	"sidecars":     true,
	"stepTemplate": true,
}

// isTektonResource checks if the document is a Tekton resource.
func isTektonResource(doc *parser.Document) bool {
	for _, prefix := range tektonAPIVersions {
		if strings.HasPrefix(doc.APIVersion, prefix) || doc.APIVersion == prefix {
			return true
		}
	}
	return false
}

// Validate validates a parsed YAML document and returns diagnostics.
func Validate(doc *parser.Document) []Diagnostic {
	if !isTektonResource(doc) {
		return nil
	}

	var diags []Diagnostic

	// Validate metadata
	diags = append(diags, validateMetadata(doc)...)

	// Resource-specific validation
	switch doc.Kind {
	case "Pipeline":
		diags = append(diags, validatePipeline(doc)...)
	case "Task", "ClusterTask":
		diags = append(diags, validateTask(doc)...)
	}

	return diags
}

func validateMetadata(doc *parser.Document) []Diagnostic {
	var diags []Diagnostic

	metadata := doc.Root.Get("metadata")
	if metadata == nil {
		// metadata itself is missing — point to root
		diags = append(diags, Diagnostic{
			Range:    doc.Root.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Required field 'metadata' is missing",
		})
		return diags
	}

	if metadata.Get("name") == nil {
		diags = append(diags, Diagnostic{
			Range:    metadata.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Required field 'metadata.name' is missing",
		})
	}

	return diags
}

func validatePipeline(doc *parser.Document) []Diagnostic {
	var diags []Diagnostic

	spec := doc.Root.Get("spec")
	if spec == nil {
		return diags
	}

	// Check for unknown fields
	diags = append(diags, checkUnknownFields(spec, knownPipelineSpecFields)...)

	// Validate tasks
	tasks := spec.Get("tasks")
	if tasks == nil {
		return diags
	}

	if !tasks.IsSequence() {
		diags = append(diags, Diagnostic{
			Range:    tasks.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Field 'tasks' must be an array/sequence",
		})
		return diags
	}

	if len(tasks.AsSequence()) == 0 {
		diags = append(diags, Diagnostic{
			Range:    tasks.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Pipeline must have at least one task",
		})
	}

	return diags
}

func validateTask(doc *parser.Document) []Diagnostic {
	var diags []Diagnostic

	spec := doc.Root.Get("spec")
	if spec == nil {
		return diags
	}

	// Check for unknown fields
	diags = append(diags, checkUnknownFields(spec, knownTaskSpecFields)...)

	// Validate steps
	steps := spec.Get("steps")
	if steps == nil {
		diags = append(diags, Diagnostic{
			Range:    spec.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Required field 'steps' is missing in Task spec",
		})
		return diags
	}

	if !steps.IsSequence() {
		diags = append(diags, Diagnostic{
			Range:    steps.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Field 'steps' must be an array/sequence",
		})
		return diags
	}

	if len(steps.AsSequence()) == 0 {
		diags = append(diags, Diagnostic{
			Range:    steps.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Task must have at least one step",
		})
	}

	return diags
}

func checkUnknownFields(node *parser.Node, knownFields map[string]bool) []Diagnostic {
	var diags []Diagnostic

	if !node.IsMapping() {
		return diags
	}

	for key, child := range node.MappingChildren {
		if !knownFields[key] {
			diags = append(diags, Diagnostic{
				Range:    child.Range,
				Severity: SeverityWarning,
				Source:   "tekton-lsp",
				Message:  fmt.Sprintf("Unknown field '%s' in spec", key),
			})
		}
	}

	return diags
}
