package symbols

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

func TestDocumentSymbols_Pipeline(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: build-pipeline
spec:
  params:
    - name: repo-url
      type: string
  tasks:
    - name: fetch
      taskRef:
        name: git-clone
    - name: build
      taskRef:
        name: build-task
  finally:
    - name: notify
      taskRef:
        name: send-notification
`)
	syms := DocumentSymbols(doc)
	require.NotEmpty(t, syms, "should return symbols for pipeline")

	// Root symbol should be the Pipeline
	assert.Equal(t, "build-pipeline", syms[0].Name)
	assert.Equal(t, SymbolKindObject, syms[0].Kind)

	// Should have children for params and tasks
	children := syms[0].Children
	require.NotEmpty(t, children)

	names := make([]string, len(children))
	for i, c := range children {
		names[i] = c.Name
	}
	assert.Contains(t, names, "params")
	assert.Contains(t, names, "tasks")
}

func TestDocumentSymbols_Task(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: build-task
spec:
  steps:
    - name: compile
      image: golang:1.25
    - name: test
      image: golang:1.25
`)
	syms := DocumentSymbols(doc)
	require.NotEmpty(t, syms)
	assert.Equal(t, "build-task", syms[0].Name)

	// Should have children for steps
	children := syms[0].Children
	names := make([]string, len(children))
	for i, c := range children {
		names[i] = c.Name
	}
	assert.Contains(t, names, "steps")
}

func TestDocumentSymbols_TaskItems(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: fetch
      taskRef:
        name: git-clone
    - name: build
      taskRef:
        name: build-task
`)
	syms := DocumentSymbols(doc)
	require.NotEmpty(t, syms)

	// Find the tasks symbol
	var tasksSym *Symbol
	for i := range syms[0].Children {
		if syms[0].Children[i].Name == "tasks" {
			tasksSym = &syms[0].Children[i]
			break
		}
	}
	require.NotNil(t, tasksSym, "should have a 'tasks' symbol")

	// Tasks should list individual task names
	require.Len(t, tasksSym.Children, 2)
	assert.Equal(t, "fetch", tasksSym.Children[0].Name)
	assert.Equal(t, "build", tasksSym.Children[1].Name)
}

func TestDocumentSymbols_HasRange(t *testing.T) {
	doc := parse(t, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
`)
	syms := DocumentSymbols(doc)
	require.NotEmpty(t, syms)

	// Root symbol should have a range starting at line 0
	assert.Equal(t, uint32(0), syms[0].Range.Start.Line)
}

func TestDocumentSymbols_NonTekton(t *testing.T) {
	doc := parse(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`)
	syms := DocumentSymbols(doc)
	assert.Empty(t, syms, "non-Tekton resources should return no symbols")
}
