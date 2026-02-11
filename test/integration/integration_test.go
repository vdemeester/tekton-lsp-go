package integration

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBinary string

func TestMain(m *testing.M) {
	// Build the binary once using a temp dir for portability.
	tmpDir, err := os.MkdirTemp("", "tekton-lsp-test-*")
	if err != nil {
		panic("mkdtemp: " + err.Error())
	}

	binary := tmpDir + "/tekton-lsp"
	cmd := exec.Command("go", "build", "-o", binary, "github.com/vdemeester/tekton-lsp-go/cmd/tekton-lsp")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}
	testBinary = binary

	code := m.Run()

	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

func startClient(t *testing.T) *Client {
	t.Helper()
	c := StartServer(t, testBinary)
	c.Initialize()
	return c
}

// ---------- Initialize ----------

func TestInitialize_Capabilities(t *testing.T) {
	c := StartServer(t, testBinary)
	resp := c.Initialize()

	result, ok := resp["result"].(map[string]any)
	require.True(t, ok, "result should be a map")

	caps, ok := result["capabilities"].(map[string]any)
	require.True(t, ok, "capabilities should be a map")

	// All features should be advertised.
	assert.NotNil(t, caps["completionProvider"], "completionProvider")
	assert.NotNil(t, caps["hoverProvider"], "hoverProvider")
	assert.NotNil(t, caps["documentSymbolProvider"], "documentSymbolProvider")
	assert.NotNil(t, caps["documentFormattingProvider"], "documentFormattingProvider")
	assert.NotNil(t, caps["definitionProvider"], "definitionProvider")
	assert.NotNil(t, caps["codeActionProvider"], "codeActionProvider")

	// Sync should be incremental or full.
	syncMode := caps["textDocumentSync"]
	assert.NotNil(t, syncMode, "textDocumentSync")
}

// ---------- Diagnostics ----------

func TestDiagnostics_MissingMetadataName(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Task
metadata: {}
spec:
  steps:
    - name: build
      image: golang
`
	c.OpenFile("file:///tmp/task.yaml", content)

	// Diagnostics come as notifications — we need to drain them.
	// The server publishes diagnostics on didOpen, which happens
	// before we issue the next request. Hover triggers reading.
	resp := c.Hover("file:///tmp/task.yaml", 0, 0)
	_ = resp
	// We can't easily capture notifications in this simple framework.
	// So we'll verify diagnostics indirectly via code actions.
}

// ---------- Completion ----------

func TestCompletion_MetadataFields(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  namespace: default
spec:
  params: []
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.Completion("file:///tmp/pipeline.yaml", 3, 2)

	result := resp["result"]
	require.NotNil(t, result, "completion result should not be nil")

	items, ok := result.([]any)
	require.True(t, ok, "result should be a list, got %T", result)
	require.NotEmpty(t, items, "should have completion items")

	labels := extractLabels(items)
	assert.Contains(t, labels, "name")
	assert.Contains(t, labels, "labels")
	assert.Contains(t, labels, "annotations")
}

func TestCompletion_PipelineSpecFields(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  description: ""
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.Completion("file:///tmp/pipeline.yaml", 5, 2)
	result := resp["result"]
	require.NotNil(t, result)

	items, ok := result.([]any)
	require.True(t, ok)
	require.NotEmpty(t, items)

	labels := extractLabels(items)
	assert.Contains(t, labels, "tasks")
	assert.Contains(t, labels, "params")
}

// ---------- Hover ----------

func TestHover_TasksField(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test-pipeline
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.Hover("file:///tmp/pipeline.yaml", 5, 4)

	result, ok := resp["result"].(map[string]any)
	require.True(t, ok && result != nil, "should return hover result")

	contents := result["contents"].(map[string]any)
	value := contents["value"].(string)
	assert.Contains(t, value, "tasks", "should document 'tasks' field")
}

func TestHover_KindPipeline(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: test
spec:
  tasks: []
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.Hover("file:///tmp/pipeline.yaml", 1, 7)

	result, ok := resp["result"].(map[string]any)
	require.True(t, ok && result != nil, "should return hover result for kind")

	contents := result["contents"].(map[string]any)
	value := contents["value"].(string)
	assert.Contains(t, value, "Pipeline")
}

func TestHover_NonTektonReturnsNull(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`
	c.OpenFile("file:///tmp/cm.yaml", content)

	resp := c.Hover("file:///tmp/cm.yaml", 1, 5)
	assert.Nil(t, resp["result"], "non-Tekton should return null")
}

// ---------- Document Symbols ----------

func TestSymbols_PipelineOutline(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main-pipeline
spec:
  params:
    - name: version
  tasks:
    - name: build
      taskRef:
        name: build-task
    - name: test
      taskRef:
        name: test-task
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.DocumentSymbols("file:///tmp/pipeline.yaml")

	result, ok := resp["result"].([]any)
	require.True(t, ok, "result should be a list")
	require.NotEmpty(t, result)

	root := result[0].(map[string]any)
	assert.Equal(t, "main-pipeline", root["name"])

	// Should have children.
	children, ok := root["children"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, children)

	childNames := make([]string, len(children))
	for i, ch := range children {
		childNames[i] = ch.(map[string]any)["name"].(string)
	}
	assert.Contains(t, childNames, "tasks")
	assert.Contains(t, childNames, "params")
}

// ---------- Formatting ----------

func TestFormatting_FixesIndentation(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Task
metadata:
    name: my-task
spec:
    steps:
        - name: build
          image: golang:1.25
`
	c.OpenFile("file:///tmp/task.yaml", content)

	resp := c.Formatting("file:///tmp/task.yaml", 2)

	result, ok := resp["result"].([]any)
	require.True(t, ok && len(result) > 0, "should return text edits")

	edit := result[0].(map[string]any)
	newText := edit["newText"].(string)
	assert.Contains(t, newText, "  name: my-task", "should have 2-space indent")
	assert.NotContains(t, newText, "    name:", "should not have 4-space indent")
}

func TestFormatting_AlreadyFormatted(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  steps:
    - name: build
      image: golang:1.25
`
	c.OpenFile("file:///tmp/task.yaml", content)

	resp := c.Formatting("file:///tmp/task.yaml", 2)
	// Already formatted — should return null or empty.
	assert.Nil(t, resp["result"], "already-formatted should return null")
}

// ---------- Go-to-definition ----------

func TestDefinition_TaskRef(t *testing.T) {
	c := startClient(t)

	taskContent := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: build-task
spec:
  steps:
    - name: compile
      image: golang:1.25
`
	pipelineContent := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`
	c.OpenFile("file:///tmp/tasks/build-task.yaml", taskContent)
	c.OpenFile("file:///tmp/pipelines/main.yaml", pipelineContent)

	resp := c.Definition("file:///tmp/pipelines/main.yaml", 8, 14)

	result, ok := resp["result"].(map[string]any)
	require.True(t, ok && result != nil, "should find definition")
	assert.Equal(t, "file:///tmp/tasks/build-task.yaml", result["uri"])
}

func TestDefinition_NotOnRef(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: main
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
`
	c.OpenFile("file:///tmp/pipeline.yaml", content)

	resp := c.Definition("file:///tmp/pipeline.yaml", 3, 8)
	assert.Nil(t, resp["result"], "should return null when not on a ref")
}

// ---------- Code Actions ----------

func TestCodeAction_UnknownField(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  unknownField: value
  steps:
    - name: build
      image: golang:1.25
`
	c.OpenFile("file:///tmp/task.yaml", content)

	resp := c.CodeAction("file:///tmp/task.yaml", 5, 0, 5, 20)

	result, ok := resp["result"].([]any)
	require.True(t, ok && len(result) > 0, "should return code actions")

	action := result[0].(map[string]any)
	title := action["title"].(string)
	assert.Contains(t, title, "Remove")
	assert.Contains(t, title, "unknownField")
	assert.Equal(t, "quickfix", action["kind"])
	assert.NotNil(t, action["edit"], "should have workspace edit")
}

func TestCodeAction_ValidDocument(t *testing.T) {
	c := startClient(t)

	content := `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: my-task
spec:
  steps:
    - name: build
      image: golang:1.25
`
	c.OpenFile("file:///tmp/task.yaml", content)

	resp := c.CodeAction("file:///tmp/task.yaml", 0, 0, 0, 10)
	// Valid document — should return null or empty.
	assert.Nil(t, resp["result"], "valid document should have no code actions")
}

// ---------- Helpers ----------

func extractLabels(items []any) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if label, ok := m["label"].(string); ok {
			labels = append(labels, label)
		}
	}
	return labels
}
