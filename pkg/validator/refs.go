package validator

import (
	"fmt"
	"regexp"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

// paramRefRe matches $(params.NAME) references.
var paramRefRe = regexp.MustCompile(`\$\(params\.([a-zA-Z_][\w-]*)\)`)

// collectDeclaredParams extracts declared parameter names from a spec.
func collectDeclaredParams(spec *parser.Node) map[string]bool {
	params := make(map[string]bool)
	if spec == nil {
		return params
	}
	p := spec.Get("params")
	if p == nil || !p.IsSequence() {
		return params
	}
	for _, item := range p.AsSequence() {
		if n := item.Get("name"); n != nil {
			params[n.AsScalar()] = true
		}
	}
	return params
}

// findParamRefs scans a node tree for $(params.NAME) references
// and returns warnings for undeclared ones.
func findParamRefs(node *parser.Node, declared map[string]bool) []Diagnostic {
	if node == nil {
		return nil
	}

	var diags []Diagnostic

	if node.IsScalar() {
		matches := paramRefRe.FindAllStringSubmatch(node.AsScalar(), -1)
		for _, m := range matches {
			name := m[1]
			if !declared[name] {
				diags = append(diags, Diagnostic{
					Range:    node.Range,
					Severity: SeverityWarning,
					Source:   "tekton-lsp",
					Message:  fmt.Sprintf("Reference to undeclared parameter '%s'", name),
				})
			}
		}
		return diags
	}

	if node.IsMapping() {
		for _, child := range node.MappingChildren {
			diags = append(diags, findParamRefs(child, declared)...)
		}
	}
	if node.IsSequence() {
		for _, child := range node.SequenceChildren {
			diags = append(diags, findParamRefs(child, declared)...)
		}
	}
	return diags
}

// validateStepImages checks that every step has an image field.
func validateStepImages(steps *parser.Node) []Diagnostic {
	if steps == nil || !steps.IsSequence() {
		return nil
	}

	var diags []Diagnostic
	for _, step := range steps.AsSequence() {
		if step.Get("image") == nil {
			// Get step name for better diagnostics.
			stepName := "unnamed"
			if n := step.Get("name"); n != nil {
				stepName = n.AsScalar()
			}
			diags = append(diags, Diagnostic{
				Range:    step.Range,
				Severity: SeverityError,
				Source:   "tekton-lsp",
				Message:  fmt.Sprintf("Step '%s' is missing required field 'image'", stepName),
			})
		}
	}
	return diags
}

// validatePipelineTasks checks taskRef.name and duplicate task names.
func validatePipelineTasks(tasks *parser.Node) []Diagnostic {
	if tasks == nil || !tasks.IsSequence() {
		return nil
	}

	var diags []Diagnostic
	seen := make(map[string]*parser.Node)

	for _, task := range tasks.AsSequence() {
		// Check for duplicate task names.
		if nameNode := task.Get("name"); nameNode != nil {
			name := nameNode.AsScalar()
			if prev, ok := seen[name]; ok {
				_ = prev
				diags = append(diags, Diagnostic{
					Range:    nameNode.Range,
					Severity: SeverityWarning,
					Source:   "tekton-lsp",
					Message:  fmt.Sprintf("Duplicate task name '%s' in pipeline", name),
				})
			} else {
				seen[name] = nameNode
			}
		}

		// Check taskRef has name.
		if taskRef := task.Get("taskRef"); taskRef != nil {
			if taskRef.Get("name") == nil {
				diags = append(diags, Diagnostic{
					Range:    taskRef.Range,
					Severity: SeverityError,
					Source:   "tekton-lsp",
					Message:  "Field 'taskRef' requires a 'name' field",
				})
			}
		}
	}

	return diags
}
