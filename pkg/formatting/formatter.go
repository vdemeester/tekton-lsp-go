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

	indent := opts.IndentSize
	if indent <= 0 {
		indent = 2
	}

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
