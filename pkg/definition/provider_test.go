package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vdemeester/tekton-lsp-go/pkg/cache"
	"github.com/vdemeester/tekton-lsp-go/pkg/parser"
)

func TestGotoDefinition_TaskRef(t *testing.T) {
	c := cache.New()

	// Index a Task
	c.Insert("file:///workspace/tasks/build.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: build-task
spec:
  steps:
    - name: build
      image: golang:1.25
`)

	// Pipeline referencing the task
	c.Insert("file:///workspace/pipeline.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`)

	pipeline, _ := c.GetParsed("file:///workspace/pipeline.yaml")

	// Position on "build-task" in taskRef.name (line 8, col 14)
	result := GotoDefinition(pipeline, parser.Position{Line: 8, Character: 14}, c)
	require.NotNil(t, result, "should find definition for taskRef")
	assert.Equal(t, "file:///workspace/tasks/build.yaml", result.URI)
}

func TestGotoDefinition_NotOnRef(t *testing.T) {
	c := cache.New()

	c.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`)

	doc, _ := c.GetParsed("file:///test.yaml")

	// Position on "main" (not a reference)
	result := GotoDefinition(doc, parser.Position{Line: 3, Character: 8}, c)
	assert.Nil(t, result, "should not find definition outside of refs")
}

func TestGotoDefinition_TaskNotFound(t *testing.T) {
	c := cache.New()

	c.Insert("file:///test.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main
spec:
  tasks:
    - name: build
      taskRef:
        name: nonexistent-task
`)

	doc, _ := c.GetParsed("file:///test.yaml")

	result := GotoDefinition(doc, parser.Position{Line: 8, Character: 14}, c)
	assert.Nil(t, result, "should not find definition for nonexistent task")
}

func TestGotoDefinition_PipelineRef(t *testing.T) {
	c := cache.New()

	// Index a Pipeline
	c.Insert("file:///workspace/pipelines/build.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: build-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: some-task
`)

	// PipelineRun referencing the pipeline
	c.Insert("file:///workspace/run.yaml", "yaml", 1, `apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  name: run-1
spec:
  pipelineRef:
    name: build-pipeline
`)

	doc, _ := c.GetParsed("file:///workspace/run.yaml")

	// Position on "build-pipeline" in pipelineRef.name (line 6, col 10)
	result := GotoDefinition(doc, parser.Position{Line: 6, Character: 10}, c)
	require.NotNil(t, result, "should find definition for pipelineRef")
	assert.Equal(t, "file:///workspace/pipelines/build.yaml", result.URI)
}
