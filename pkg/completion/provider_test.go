package completion

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

func completionLabels(items []CompletionItem) []string {
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label
	}
	return labels
}

func TestComplete_PipelineSpec(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  
`)
	// Position inside spec (line 5, col 2)
	items := Complete(doc, parser.Position{Line: 5, Character: 2})
	require.NotEmpty(t, items, "should offer completions in Pipeline spec")

	labels := completionLabels(items)
	assert.Contains(t, labels, "tasks")
	assert.Contains(t, labels, "params")
	assert.Contains(t, labels, "workspaces")
	assert.Contains(t, labels, "finally")
}

func TestComplete_TaskSpec(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  
`)
	items := Complete(doc, parser.Position{Line: 5, Character: 2})
	require.NotEmpty(t, items, "should offer completions in Task spec")

	labels := completionLabels(items)
	assert.Contains(t, labels, "steps")
	assert.Contains(t, labels, "params")
	assert.Contains(t, labels, "results")
}

func TestComplete_Metadata(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  
`)
	items := Complete(doc, parser.Position{Line: 3, Character: 2})
	require.NotEmpty(t, items, "should offer completions in metadata")

	labels := completionLabels(items)
	assert.Contains(t, labels, "name")
	assert.Contains(t, labels, "namespace")
	assert.Contains(t, labels, "labels")
}

func TestComplete_PipelineTask(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - 
`)
	// Position inside a pipeline task item (line 6, col 6)
	items := Complete(doc, parser.Position{Line: 6, Character: 6})
	require.NotEmpty(t, items, "should offer completions for pipeline task")

	labels := completionLabels(items)
	assert.Contains(t, labels, "name")
	assert.Contains(t, labels, "taskRef")
	assert.Contains(t, labels, "runAfter")
}

func TestComplete_Step(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - 
`)
	items := Complete(doc, parser.Position{Line: 6, Character: 6})
	require.NotEmpty(t, items, "should offer completions for step")

	labels := completionLabels(items)
	assert.Contains(t, labels, "name")
	assert.Contains(t, labels, "image")
	assert.Contains(t, labels, "script")
	assert.Contains(t, labels, "command")
}

func TestComplete_OutsideSpec(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
`)
	// Position outside any relevant context (line 0, col 0)
	items := Complete(doc, parser.Position{Line: 0, Character: 0})
	assert.Empty(t, items, "should offer no completions at root level")
}

func TestComplete_NonTekton(t *testing.T) {
	doc := parse(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  
`)
	items := Complete(doc, parser.Position{Line: 5, Character: 2})
	assert.Empty(t, items, "should offer no completions for non-Tekton resources")
}

func TestCompletionItem_HasDetail(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  
`)
	items := Complete(doc, parser.Position{Line: 5, Character: 2})
	require.NotEmpty(t, items)

	// All items should have a detail/description
	for _, item := range items {
		assert.NotEmpty(t, item.Detail, "completion item '%s' should have detail", item.Label)
	}
}
