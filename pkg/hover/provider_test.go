package hover

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

func TestHover_APIVersion(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
`)
	result := Hover(doc, parser.Position{Line: 0, Character: 3})
	require.NotNil(t, result, "should return hover for apiVersion")
	assert.Contains(t, result.Content, "apiVersion")
}

func TestHover_Kind(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
`)
	result := Hover(doc, parser.Position{Line: 1, Character: 3})
	require.NotNil(t, result, "should return hover for kind")
	assert.Contains(t, result.Content, "Pipeline")
}

func TestHover_MetadataName(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: my-pipeline
`)
	result := Hover(doc, parser.Position{Line: 3, Character: 4})
	require.NotNil(t, result, "should return hover for metadata.name")
	assert.Contains(t, result.Content, "name")
}

func TestHover_Tasks(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`)
	result := Hover(doc, parser.Position{Line: 5, Character: 4})
	require.NotNil(t, result, "should return hover for tasks")
	assert.Contains(t, result.Content, "tasks")
}

func TestHover_Steps(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`)
	result := Hover(doc, parser.Position{Line: 5, Character: 4})
	require.NotNil(t, result, "should return hover for steps")
	assert.Contains(t, result.Content, "steps")
}

func TestHover_HasRange(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
`)
	result := Hover(doc, parser.Position{Line: 1, Character: 3})
	require.NotNil(t, result)
	assert.NotNil(t, result.Range, "hover result should include range")
}

func TestHover_OutsideDocument(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
`)
	result := Hover(doc, parser.Position{Line: 100, Character: 0})
	assert.Nil(t, result, "should return nil for position outside document")
}

func TestHover_NonTekton(t *testing.T) {
	doc := parse(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`)
	result := Hover(doc, parser.Position{Line: 1, Character: 3})
	assert.Nil(t, result, "should return nil for non-Tekton resources")
}
