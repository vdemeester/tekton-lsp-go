package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

func parse(t *testing.T, yaml string) *parser.Document {
	t.Helper()
	doc, err := parser.ParseYAML("test.yaml", yaml)
	require.NoError(t, err)
	return doc
}

func TestValidate_ValidPipeline_NoDiagnostics(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: build-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)
	diags := Validate(doc)
	assert.Empty(t, diags, "valid pipeline should produce no diagnostics")
}

func TestValidate_ValidTask_NoDiagnostics(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: build-task
spec:
  steps:
    - name: build
      image: golang:1.25
      script: |
        go build ./...
`)
	diags := Validate(doc)
	assert.Empty(t, diags, "valid task should produce no diagnostics")
}

func TestValidate_MissingMetadataName(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  namespace: default
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)
	diags := Validate(doc)
	require.Len(t, diags, 1)
	assert.Equal(t, SeverityError, diags[0].Severity)
	assert.Contains(t, diags[0].Message, "metadata.name")
}

func TestValidate_MissingMetadata(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)
	diags := Validate(doc)
	require.NotEmpty(t, diags)

	hasMetadataError := false
	for _, d := range diags {
		if d.Severity == SeverityError && (contains(d.Message, "metadata") || contains(d.Message, "required")) {
			hasMetadataError = true
		}
	}
	assert.True(t, hasMetadataError, "should report missing metadata")
}

func TestValidate_Pipeline_EmptyTasks(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks: []
`)
	diags := Validate(doc)
	require.NotEmpty(t, diags)
	assert.Equal(t, SeverityError, diags[0].Severity)
	assert.Contains(t, diags[0].Message, "at least one task")
}

func TestValidate_Pipeline_TasksWrongType(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks: "should-be-array"
`)
	diags := Validate(doc)
	require.NotEmpty(t, diags)

	hasTypeError := false
	for _, d := range diags {
		if contains(d.Message, "array") || contains(d.Message, "sequence") {
			hasTypeError = true
		}
	}
	assert.True(t, hasTypeError, "should report tasks type error")
}

func TestValidate_Pipeline_UnknownSpecField(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  taskz:
    - name: build
      taskRef:
        name: some-task
`)
	diags := Validate(doc)

	warnings := filterBySeverity(diags, SeverityWarning)
	require.NotEmpty(t, warnings, "should have at least one warning")
	assert.Contains(t, warnings[0].Message, "taskz")
}

func TestValidate_Pipeline_MultipleErrors(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  namespace: default
spec:
  tasks: []
`)
	diags := Validate(doc)
	assert.GreaterOrEqual(t, len(diags), 2, "should report both missing name and empty tasks")
}

func TestValidate_Task_MissingSteps(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  params:
    - name: foo
`)
	diags := Validate(doc)

	hasStepsError := false
	for _, d := range diags {
		if d.Severity == SeverityError && contains(d.Message, "steps") {
			hasStepsError = true
		}
	}
	assert.True(t, hasStepsError, "should report missing steps in Task")
}

func TestValidate_Task_EmptySteps(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  steps: []
`)
	diags := Validate(doc)
	require.NotEmpty(t, diags)
	assert.Equal(t, SeverityError, diags[0].Severity)
	assert.Contains(t, diags[0].Message, "at least one step")
}

func TestValidate_NonTektonResource_NoDiagnostics(t *testing.T) {
	doc := parse(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
`)
	diags := Validate(doc)
	assert.Empty(t, diags, "non-Tekton resources should not be validated")
}

func TestDiagnostic_HasAccuratePosition(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  namespace: default
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)
	diags := Validate(doc)
	require.NotEmpty(t, diags)

	// The diagnostic for missing metadata.name should point to the metadata node
	d := diags[0]
	assert.Equal(t, uint32(2), d.Range.Start.Line, "diagnostic should point to metadata line")
}

// helpers

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func filterBySeverity(diags []Diagnostic, severity Severity) []Diagnostic {
	var result []Diagnostic
	for _, d := range diags {
		if d.Severity == severity {
			result = append(result, d)
		}
	}
	return result
}
