package hover

// docs maps field keys to markdown documentation.
var docs = map[string]string{
	// Top-level
	"apiVersion": "**apiVersion** — The API version for this resource.\n\nTekton resources use `tekton.dev/v1` or `tekton.dev/v1beta1`.",
	"kind":       "**kind** — The type of Tekton resource.\n\nCommon kinds: `Pipeline`, `Task`, `PipelineRun`, `TaskRun`.",
	"metadata":   "**metadata** — Standard Kubernetes metadata.\n\nIncludes `name`, `namespace`, `labels`, and `annotations`.",

	// Metadata
	"name":        "**name** — The name of the resource (required).\n\nMust be unique within the namespace.",
	"namespace":   "**namespace** — The Kubernetes namespace for this resource.",
	"labels":      "**labels** — Key-value pairs for organizing resources.",
	"annotations": "**annotations** — Key-value pairs for non-identifying metadata.",

	// Pipeline spec
	"tasks":       "**tasks** — The list of Tasks to execute in the Pipeline.\n\nEach task references a `Task` resource via `taskRef` or defines one inline via `taskSpec`.",
	"finally":     "**finally** — Tasks that run after all other tasks complete.\n\nThese tasks run regardless of whether previous tasks succeeded or failed.",
	"params":      "**params** — Parameters accepted by this resource.\n\nParameters can be referenced using `$(params.name)` syntax.",
	"workspaces":  "**workspaces** — Workspace declarations for shared storage.\n\nWorkspaces provide a way to share data between tasks.",
	"results":     "**results** — Results produced by this resource.\n\nResults can be referenced by subsequent tasks using `$(tasks.taskName.results.resultName)`.",
	"description": "**description** — A human-readable description of this resource.",

	// Pipeline task
	"taskRef":  "**taskRef** — Reference to an existing Task resource.\n\nSpecify `name` (and optionally `kind` for ClusterTask).",
	"taskSpec": "**taskSpec** — Inline Task specification.\n\nDefine a Task directly within the Pipeline instead of referencing an existing one.",
	"runAfter": "**runAfter** — List of task names that must complete before this task runs.\n\nUsed to define execution order.",
	"when":     "**when** — Conditional expressions that determine if this task runs.\n\nUse `input`, `operator`, and `values` fields.",
	"matrix":   "**matrix** — Matrix parameters for fan-out execution.\n\nRuns the task multiple times with different parameter combinations.",

	// Task spec
	"steps":        "**steps** — The sequence of containers to run in this Task.\n\nEach step runs in order within a shared Pod.",
	"volumes":      "**volumes** — Kubernetes volumes to make available to steps.",
	"sidecars":     "**sidecars** — Sidecar containers that run alongside steps.\n\nUseful for services like databases needed during execution.",
	"stepTemplate": "**stepTemplate** — Default values applied to all steps.\n\nSet common image, env vars, or resource limits.",

	// Step
	"image":      "**image** — The container image to use for this step (required).\n\nExample: `golang:1.25`, `ubuntu:latest`.",
	"script":     "**script** — A script to execute in the container.\n\nThe script is written to a file and executed. Supports shebangs.",
	"command":    "**command** — The entrypoint for the container.\n\nOverrides the image's default entrypoint.",
	"args":       "**args** — Arguments passed to the command.",
	"env":        "**env** — Environment variables for the container.\n\nEach entry has `name` and `value` (or `valueFrom`).",
	"workingDir": "**workingDir** — The working directory for the container.",
}

// getDocumentation returns markdown documentation for a field key.
func getDocumentation(key string) (string, bool) {
	doc, ok := docs[key]
	return doc, ok
}
