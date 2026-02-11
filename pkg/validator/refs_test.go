package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -- param ref validation --

func TestValidate_ParamRefToUndeclaredParam(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  steps:
    - name: build
      image: golang:1.25
      script: |
        echo $(params.missing-param)
`)
	diags := Validate(doc)
	warnings := filterBySeverity(diags, SeverityWarning)
	require.NotEmpty(t, warnings, "should warn about undeclared param ref")
	assert.Contains(t, warnings[0].Message, "missing-param")
}

func TestValidate_ParamRefToExistingParam(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  params:
    - name: version
  steps:
    - name: build
      image: golang:1.25
      script: |
        echo $(params.version)
`)
	diags := Validate(doc)
	warnings := filterBySeverity(diags, SeverityWarning)
	assert.Empty(t, warnings, "declared param should not produce warning")
}

func TestValidate_Pipeline_ParamRefInTaskParams(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
      params:
        - name: repo
          value: $(params.repo-url)
`)
	diags := Validate(doc)
	warnings := filterBySeverity(diags, SeverityWarning)
	require.NotEmpty(t, warnings, "should warn about undeclared pipeline param")
	assert.Contains(t, warnings[0].Message, "repo-url")
}

func TestValidate_Pipeline_DeclaredParamRef(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  params:
    - name: repo-url
  tasks:
    - name: build
      taskRef:
        name: build-task
      params:
        - name: repo
          value: $(params.repo-url)
`)
	diags := Validate(doc)
	warnings := filterBySeverity(diags, SeverityWarning)
	assert.Empty(t, warnings)
}

// -- step image required --

func TestValidate_Task_StepMissingImage(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  steps:
    - name: build
      script: echo hello
`)
	diags := Validate(doc)
	errors := filterBySeverity(diags, SeverityError)
	hasImageError := false
	for _, d := range errors {
		if contains(d.Message, "image") {
			hasImageError = true
		}
	}
	assert.True(t, hasImageError, "step without image should produce error")
}

func TestValidate_Task_StepWithImage(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test-task
spec:
  steps:
    - name: build
      image: golang:1.25
`)
	diags := Validate(doc)
	assert.Empty(t, diags)
}

// -- taskRef.name required --

func TestValidate_Pipeline_TaskRefMissingName(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef: {}
`)
	diags := Validate(doc)
	errors := filterBySeverity(diags, SeverityError)
	hasRefError := false
	for _, d := range errors {
		if contains(d.Message, "taskRef") && contains(d.Message, "name") {
			hasRefError = true
		}
	}
	assert.True(t, hasRefError, "taskRef without name should produce error")
}

// -- duplicate task names --

func TestValidate_Pipeline_DuplicateTaskNames(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
    - name: build
      taskRef:
        name: other-task
`)
	diags := Validate(doc)
	warnings := filterBySeverity(diags, SeverityWarning)
	hasDupWarning := false
	for _, d := range warnings {
		if contains(d.Message, "Duplicate") && contains(d.Message, "build") {
			hasDupWarning = true
		}
	}
	assert.True(t, hasDupWarning, "duplicate task names should produce warning")
}

func TestValidate_Pipeline_UniqueTaskNames(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
    - name: test
      taskRef:
        name: test-task
`)
	diags := Validate(doc)
	assert.Empty(t, diags)
}
