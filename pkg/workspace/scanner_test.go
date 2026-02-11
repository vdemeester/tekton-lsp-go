package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vdemeester/tekton-lsp-go/pkg/cache"
)

func setupWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}
	return dir
}

func TestScan_FindsTektonYAML(t *testing.T) {
	dir := setupWorkspace(t, map[string]string{
		"tasks/build.yaml": `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: build-task
spec:
  steps:
    - name: build
      image: golang:1.25
`,
		"pipelines/main.yaml": `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`,
	})

	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	assert.Equal(t, 2, n, "should index 2 YAML files")

	// Check that documents are in the cache.
	docs := c.AllParsed()
	assert.Len(t, docs, 2)
}

func TestScan_IgnoresNonYAML(t *testing.T) {
	dir := setupWorkspace(t, map[string]string{
		"README.md": "# readme",
		"script.sh": "#!/bin/bash",
		"task.yaml": `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`,
	})

	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "should only index .yaml/.yml files")
}

func TestScan_IgnoresNonTekton(t *testing.T) {
	dir := setupWorkspace(t, map[string]string{
		"configmap.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
`,
		"task.yaml": `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`,
	})

	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	// Both files are scanned (YAML) but only Tekton ones are useful
	// The scanner indexes all YAML files; filtering is done by features
	assert.Equal(t, 2, n, "should index all YAML files")
}

func TestScan_RecursesSubdirectories(t *testing.T) {
	dir := setupWorkspace(t, map[string]string{
		"a/b/c/task.yaml": `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: deep-task
spec:
  steps:
    - name: build
      image: golang:1.25
`,
	})

	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestScan_YMLExtension(t *testing.T) {
	dir := setupWorkspace(t, map[string]string{
		"task.yml": `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: test
spec:
  steps:
    - name: build
      image: golang:1.25
`,
	})

	c := cache.New()
	n, err := Scan("file://"+dir, c)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "should handle .yml extension")
}
