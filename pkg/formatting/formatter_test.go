package formatting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat_FixesIndentation(t *testing.T) {
	input := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
    name: test
spec:
    tasks:
        - name: build
          taskRef:
              name: build-task
`
	expected := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`
	result, err := Format(input, Options{IndentSize: 2})
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestFormat_PreservesValidYAML(t *testing.T) {
	input := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`
	result, err := Format(input, Options{IndentSize: 2})
	require.NoError(t, err)
	assert.Equal(t, input, result, "already-formatted YAML should not change")
}

func TestFormat_EmptyContent(t *testing.T) {
	result, err := Format("", Options{IndentSize: 2})
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestFormat_DefaultOptions(t *testing.T) {
	input := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
    name: test
`
	result, err := Format(input, Options{})
	require.NoError(t, err)
	// Default indent should be 2
	assert.Contains(t, result, "  name: test")
}

func TestFormat_LargeTabSize_ClampedToTwo(t *testing.T) {
	input := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`
	// Even when editor sends tabSize=4 or 8, YAML should stay at 2
	result, err := Format(input, Options{IndentSize: 4})
	require.NoError(t, err)
	assert.Contains(t, result, "  name: test", "should use 2-space indent regardless of tabSize")
	assert.NotContains(t, result, "    name:", "should not use 4-space indent")
}
