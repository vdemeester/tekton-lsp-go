package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAML_SimplePipeline(t *testing.T) {
	content, err := os.ReadFile("../../test/fixtures/pipeline.yaml")
	require.NoError(t, err)

	doc, err := ParseYAML("pipeline.yaml", string(content))
	require.NoError(t, err)
	assert.NotNil(t, doc)

	assert.Equal(t, "tekton.dev/v1", doc.APIVersion)
	assert.Equal(t, "Pipeline", doc.Kind)
}

func TestParseYAML_SimpleTask(t *testing.T) {
	content, err := os.ReadFile("../../test/fixtures/task.yaml")
	require.NoError(t, err)

	doc, err := ParseYAML("task.yaml", string(content))
	require.NoError(t, err)
	assert.NotNil(t, doc)

	assert.Equal(t, "tekton.dev/v1", doc.APIVersion)
	assert.Equal(t, "Task", doc.Kind)
}

func TestParseYAML_InlineContent(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: hello
`
	doc, err := ParseYAML("test.yaml", yaml)
	require.NoError(t, err)

	assert.Equal(t, "tekton.dev/v1", doc.APIVersion)
	assert.Equal(t, "Task", doc.Kind)
}

func TestParseYAML_EmptyContent(t *testing.T) {
	_, err := ParseYAML("empty.yaml", "")
	assert.Error(t, err, "parsing empty content should return an error")
}

func TestNode_GetChild(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test-pipeline
  namespace: default
`
	doc, err := ParseYAML("test.yaml", yaml)
	require.NoError(t, err)

	// Root should be a mapping
	assert.True(t, doc.Root.IsMapping(), "root should be a mapping")

	// Should be able to get children by key
	metadata := doc.Root.Get("metadata")
	require.NotNil(t, metadata, "metadata should exist")
	assert.True(t, metadata.IsMapping(), "metadata should be a mapping")

	// Should get nested children
	name := metadata.Get("name")
	require.NotNil(t, name, "metadata.name should exist")
	assert.True(t, name.IsScalar(), "name should be a scalar")
	assert.Equal(t, "test-pipeline", name.AsScalar())

	// Non-existent key returns nil
	missing := doc.Root.Get("nonexistent")
	assert.Nil(t, missing, "non-existent key should return nil")
}

func TestNode_Sequence(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Pipeline
spec:
  tasks:
    - name: task1
      taskRef:
        name: build
    - name: task2
      taskRef:
        name: test
`
	doc, err := ParseYAML("test.yaml", yaml)
	require.NoError(t, err)

	spec := doc.Root.Get("spec")
	require.NotNil(t, spec)

	tasks := spec.Get("tasks")
	require.NotNil(t, tasks)
	assert.True(t, tasks.IsSequence(), "tasks should be a sequence")

	items := tasks.AsSequence()
	require.Len(t, items, 2, "should have 2 tasks")

	// First task
	assert.True(t, items[0].IsMapping())
	taskName := items[0].Get("name")
	require.NotNil(t, taskName)
	assert.Equal(t, "task1", taskName.AsScalar())

	// Second task
	taskRef := items[1].Get("taskRef")
	require.NotNil(t, taskRef)
	refName := taskRef.Get("name")
	require.NotNil(t, refName)
	assert.Equal(t, "test", refName.AsScalar())
}

func TestNode_PositionTracking(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test-pipeline
  namespace: default
spec:
  tasks:
    - name: task1
`
	doc, err := ParseYAML("test.yaml", yaml)
	require.NoError(t, err)

	// Root starts at line 0
	assert.Equal(t, uint32(0), doc.Root.Range.Start.Line, "root should start at line 0")

	// metadata starts at line 2
	metadata := doc.Root.Get("metadata")
	require.NotNil(t, metadata)
	assert.Equal(t, uint32(2), metadata.Range.Start.Line, "metadata should start at line 2")

	// spec starts at line 5
	spec := doc.Root.Get("spec")
	require.NotNil(t, spec)
	assert.Equal(t, uint32(5), spec.Range.Start.Line, "spec should start at line 5")

	// Positions should have real column values (not zero for all)
	assert.True(t, metadata.Range.End.Character > 0, "should have real end character position")
}

func TestDocument_FindNodeAtPosition(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test-pipeline
`
	doc, err := ParseYAML("test.yaml", yaml)
	require.NoError(t, err)

	// Position on "kind" line (line 1, column 6 = inside "Pipeline")
	node := doc.FindNodeAtPosition(Position{Line: 1, Character: 6})
	require.NotNil(t, node, "should find a node at line 1, col 6")

	// Position on "name" value (line 3, column 10 = inside "test-pipeline")
	node = doc.FindNodeAtPosition(Position{Line: 3, Character: 10})
	require.NotNil(t, node, "should find a node at line 3, col 10")

	// Position outside document
	node = doc.FindNodeAtPosition(Position{Line: 100, Character: 0})
	assert.Nil(t, node, "should not find a node outside the document")
}

func TestParseAllYAML_MultiDocument(t *testing.T) {
	yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
---
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  steps:
    - name: build
      image: golang:1.25
---
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: my-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: my-task
`
	docs, err := ParseAllYAML("multi.yaml", yaml)
	require.NoError(t, err)
	require.Len(t, docs, 3, "should parse 3 documents")

	// First doc: ConfigMap
	assert.Equal(t, "v1", docs[0].APIVersion)
	assert.Equal(t, "ConfigMap", docs[0].Kind)
	assert.Equal(t, 0, docs[0].Index)

	// Second doc: Task
	assert.Equal(t, "tekton.dev/v1", docs[1].APIVersion)
	assert.Equal(t, "Task", docs[1].Kind)
	assert.Equal(t, 1, docs[1].Index)
	taskName := docs[1].Root.Get("metadata").Get("name")
	require.NotNil(t, taskName)
	assert.Equal(t, "my-task", taskName.AsScalar())

	// Third doc: Pipeline
	assert.Equal(t, "tekton.dev/v1", docs[2].APIVersion)
	assert.Equal(t, "Pipeline", docs[2].Kind)
	assert.Equal(t, 2, docs[2].Index)

	// Verify positions are correct (Task starts after ---, around line 7)
	assert.True(t, docs[1].Root.Range.Start.Line >= 7, "Task doc should start after line 7, got %d", docs[1].Root.Range.Start.Line)
	assert.True(t, docs[2].Root.Range.Start.Line >= 16, "Pipeline doc should start after line 16, got %d", docs[2].Root.Range.Start.Line)
}

func TestParseAllYAML_SingleDocument(t *testing.T) {
	yaml := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: hello
`
	docs, err := ParseAllYAML("single.yaml", yaml)
	require.NoError(t, err)
	require.Len(t, docs, 1, "single doc file should return 1 document")
	assert.Equal(t, "Task", docs[0].Kind)
	assert.Equal(t, 0, docs[0].Index)
}

func TestParseAllYAML_FindNodeInSecondDocument(t *testing.T) {
	yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
---
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  steps:
    - name: build
      image: golang:1.25
`
	docs, err := ParseAllYAML("multi.yaml", yaml)
	require.NoError(t, err)
	require.Len(t, docs, 2)

	// Find node in second document (Task) — "spec" line
	specNode := docs[1].FindNodeAtPosition(Position{Line: 9, Character: 0})
	require.NotNil(t, specNode, "should find node in second document")

	// "steps" should be findable
	steps := docs[1].Root.Get("spec").Get("steps")
	require.NotNil(t, steps, "should find steps in second document")
	assert.True(t, steps.IsSequence())
	assert.Len(t, steps.AsSequence(), 1)
}
