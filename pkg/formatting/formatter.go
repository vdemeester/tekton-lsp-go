package formatting

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// Options controls formatting behavior.
type Options struct {
	IndentSize int
}

// Format reformats YAML content with consistent indentation.
func Format(content string, opts Options) (string, error) {
	if content == "" {
		return "", nil
	}

	// YAML convention is 2-space indent. Always use 2 regardless of
	// what the editor sends (tabSize can be 4 or 8 in many editors).
	indent := 2
	_ = opts.IndentSize // acknowledged but overridden

	// Parse then re-serialize with consistent formatting.
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(&node); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
