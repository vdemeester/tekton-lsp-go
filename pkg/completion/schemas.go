package completion

// FieldType represents the type of a Tekton schema field.
type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeArray
	FieldTypeObject
	FieldTypeBoolean
)

// FieldSchema defines a completable field with its metadata.
type FieldSchema struct {
	Name        string
	Description string
	Type        FieldType
	Required    bool
}

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label  string
	Detail string
	Kind   FieldType
}

var metadataFields = []FieldSchema{
	{Name: "name", Description: "Resource name (required)", Type: FieldTypeString, Required: true},
	{Name: "namespace", Description: "Resource namespace", Type: FieldTypeString},
	{Name: "labels", Description: "Resource labels", Type: FieldTypeObject},
	{Name: "annotations", Description: "Resource annotations", Type: FieldTypeObject},
}

var pipelineSpecFields = []FieldSchema{
	{Name: "tasks", Description: "Pipeline tasks to execute", Type: FieldTypeArray, Required: true},
	{Name: "finally", Description: "Tasks to run after all other tasks", Type: FieldTypeArray},
	{Name: "params", Description: "Pipeline parameters", Type: FieldTypeArray},
	{Name: "workspaces", Description: "Pipeline workspaces", Type: FieldTypeArray},
	{Name: "results", Description: "Pipeline results", Type: FieldTypeArray},
	{Name: "description", Description: "Pipeline description", Type: FieldTypeString},
}

var pipelineTaskFields = []FieldSchema{
	{Name: "name", Description: "Task name (required)", Type: FieldTypeString, Required: true},
	{Name: "taskRef", Description: "Reference to an existing Task", Type: FieldTypeObject},
	{Name: "taskSpec", Description: "Inline Task specification", Type: FieldTypeObject},
	{Name: "params", Description: "Task parameters", Type: FieldTypeArray},
	{Name: "workspaces", Description: "Workspace bindings", Type: FieldTypeArray},
	{Name: "runAfter", Description: "Tasks that must complete before this task", Type: FieldTypeArray},
	{Name: "when", Description: "Conditional execution expressions", Type: FieldTypeArray},
	{Name: "matrix", Description: "Matrix parameters for fan-out", Type: FieldTypeObject},
}

var taskSpecFields = []FieldSchema{
	{Name: "steps", Description: "Task steps to execute", Type: FieldTypeArray, Required: true},
	{Name: "params", Description: "Task parameters", Type: FieldTypeArray},
	{Name: "workspaces", Description: "Task workspaces", Type: FieldTypeArray},
	{Name: "results", Description: "Task results", Type: FieldTypeArray},
	{Name: "volumes", Description: "Kubernetes volumes", Type: FieldTypeArray},
	{Name: "sidecars", Description: "Sidecar containers", Type: FieldTypeArray},
	{Name: "stepTemplate", Description: "Template for step defaults", Type: FieldTypeObject},
	{Name: "description", Description: "Task description", Type: FieldTypeString},
}

var stepFields = []FieldSchema{
	{Name: "name", Description: "Step name (required)", Type: FieldTypeString, Required: true},
	{Name: "image", Description: "Container image (required)", Type: FieldTypeString, Required: true},
	{Name: "script", Description: "Script to execute", Type: FieldTypeString},
	{Name: "command", Description: "Container entrypoint", Type: FieldTypeArray},
	{Name: "args", Description: "Container arguments", Type: FieldTypeArray},
	{Name: "env", Description: "Environment variables", Type: FieldTypeArray},
	{Name: "workingDir", Description: "Working directory", Type: FieldTypeString},
}
