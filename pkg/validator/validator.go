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

// knownParamFields are the valid fields in a param definition.
var knownParamFields = map[string]bool{
	"name":        true,
	"type":        true,
	"description": true,
	"default":     true,
	"properties":  true,
	"enum":        true,
}

// knownStepFields are the valid fields in a step definition.
var knownStepFields = map[string]bool{
	"name":            true,
	"image":           true,
	"command":         true,
	"args":            true,
	"script":          true,
	"env":             true,
	"envFrom":         true,
	"workingDir":      true,
	"volumeMounts":    true,
	"resources":       true,
	"securityContext": true,
	"timeout":         true,
	"onError":         true,
	"stdoutConfig":    true,
	"stderrConfig":    true,
	"params":          true,
	"ref":             true,
	"computeResources": true,
	"workspaces":      true,
	"results":         true,
}

// knownWorkspaceFields are the valid fields in a workspace declaration.
var knownWorkspaceFields = map[string]bool{
	"name":        true,
	"description": true,
	"mountPath":   true,
	"readOnly":    true,
	"optional":    true,
}

// knownResultFields are the valid fields in a result definition.
var knownResultFields = map[string]bool{
	"name":        true,
	"type":        true,
	"description": true,
	"properties":  true,
	"value":       true,
}

// knownPipelineTaskFields are the valid fields in a pipeline task entry.
var knownPipelineTaskFields = map[string]bool{
	"name":         true,
	"taskRef":      true,
	"taskSpec":     true,
	"runAfter":     true,
	"params":       true,
	"workspaces":   true,
	"matrix":       true,
	"when":         true,
	"timeout":      true,
	"retries":      true,
	"onError":      true,
	"displayName":  true,
	"description":  true,
	"pipelineRef":  true,
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

	if metadata.Get("name") == nil && metadata.Get("generateName") == nil {
		diags = append(diags, Diagnostic{
			Range:    metadata.Range,
			Severity: SeverityError,
			Source:   "tekton-lsp",
			Message:  "Required field 'metadata.name' or 'metadata.generateName' is missing",
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

	// Validate individual tasks (taskRef.name, duplicate names)
	diags = append(diags, validatePipelineTasks(tasks)...)

	// Validate pipeline task fields
	diags = append(diags, validateSequenceItems(tasks, knownPipelineTaskFields, "pipeline task")...)

	// Validate param fields
	if params := spec.Get("params"); params != nil {
		diags = append(diags, validateSequenceItems(params, knownParamFields, "param")...)
	}

	// Validate workspace fields
	if workspaces := spec.Get("workspaces"); workspaces != nil {
		diags = append(diags, validateSequenceItems(workspaces, knownWorkspaceFields, "workspace")...)
	}

	// Validate result fields
	if results := spec.Get("results"); results != nil {
		diags = append(diags, validateSequenceItems(results, knownResultFields, "result")...)
	}

	// Validate finally task fields
	if finally := spec.Get("finally"); finally != nil && finally.IsSequence() {
		diags = append(diags, validateSequenceItems(finally, knownPipelineTaskFields, "finally task")...)
	}

	// Validate param references
	declaredParams := collectDeclaredParams(spec)
	diags = append(diags, findParamRefs(tasks, declaredParams)...)

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

	// Validate step fields
	diags = append(diags, validateStepImages(steps)...)
	diags = append(diags, validateSequenceItems(steps, knownStepFields, "step")...)

	// Validate param fields
	if params := spec.Get("params"); params != nil {
		diags = append(diags, validateSequenceItems(params, knownParamFields, "param")...)
	}

	// Validate workspace fields
	if workspaces := spec.Get("workspaces"); workspaces != nil {
		diags = append(diags, validateSequenceItems(workspaces, knownWorkspaceFields, "workspace")...)
	}

	// Validate result fields
	if results := spec.Get("results"); results != nil {
		diags = append(diags, validateSequenceItems(results, knownResultFields, "result")...)
	}

	// Validate param references
	declaredParams := collectDeclaredParams(spec)
	diags = append(diags, findParamRefs(spec, declaredParams)...)

	return diags
}

func checkUnknownFields(node *parser.Node, knownFields map[string]bool) []Diagnostic {
	return checkUnknownFieldsContext(node, knownFields, "spec")
}

func checkUnknownFieldsContext(node *parser.Node, knownFields map[string]bool, context string) []Diagnostic {
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
				Message:  fmt.Sprintf("Unknown field '%s' in %s", key, context),
			})
		}
	}

	return diags
}

// validateSequenceItems validates each item in a sequence against known fields.
func validateSequenceItems(node *parser.Node, knownFields map[string]bool, context string) []Diagnostic {
	var diags []Diagnostic
	if node == nil || !node.IsSequence() {
		return diags
	}
	for _, item := range node.AsSequence() {
		if item.IsMapping() {
			diags = append(diags, checkUnknownFieldsContext(item, knownFields, context)...)
		}
	}
	return diags
}
