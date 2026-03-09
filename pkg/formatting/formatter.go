package formatting

import (
	"bytes"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Options controls formatting behavior.
type Options struct {
	IndentSize int
}

// Format reformats YAML content with consistent indentation.
// Handles multi-document YAML files (separated by ---).
func Format(content string, opts Options) (string, error) {
	if content == "" {
		return "", nil
	}

	// YAML convention is 2-space indent. Always use 2 regardless of
	// what the editor sends (tabSize can be 4 or 8 in many editors).
	indent := 2
	_ = opts.IndentSize // acknowledged but overridden

	// Use Decoder to handle multiple YAML documents in one file.
	dec := yaml.NewDecoder(strings.NewReader(content))

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)

	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if err := enc.Encode(&node); err != nil {
			return "", err
		}
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
