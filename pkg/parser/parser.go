package parser

import (
	"fmt"

	tree_sitter_yaml "github.com/tree-sitter-grammars/tree-sitter-yaml/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// ParseYAML parses YAML content into a Document with position tracking using tree-sitter.
// Deprecated: Use ParseAllYAML for multi-document support. This returns only the first document.
func ParseYAML(filename, content string) (*Document, error) {
	docs, err := ParseAllYAML(filename, content)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents found")
	}
	return docs[0], nil
}

// ParseAllYAML parses YAML content into multiple Documents (one per --- separated document)
// with position tracking using tree-sitter.
func ParseAllYAML(filename, content string) ([]*Document, error) {
	if content == "" {
		return nil, fmt.Errorf("empty content")
	}

	parser := tree_sitter.NewParser()
	defer parser.Close()

	lang := tree_sitter.NewLanguage(tree_sitter_yaml.Language())
	if err := parser.SetLanguage(lang); err != nil {
		return nil, fmt.Errorf("failed to set language: %w", err)
	}

	tree := parser.Parse([]byte(content), nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse YAML")
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	contentBytes := []byte(content)

	return buildDocuments(rootNode, contentBytes, filename)
}

// buildDocuments extracts all YAML documents from the tree-sitter stream node.
func buildDocuments(rootNode *tree_sitter.Node, content []byte, filename string) ([]*Document, error) {
	var docs []*Document

	// tree-sitter YAML produces: stream -> document* -> block_node -> block_mapping
	// For single-doc files: stream has 1 document child.
	// For multi-doc files: stream has N document children.
	if rootNode.Kind() != "stream" {
		// Shouldn't happen, but handle gracefully.
		doc, err := buildSingleDocument(rootNode, content, filename, 0)
		if err != nil {
			return nil, err
		}
		if doc != nil {
			docs = append(docs, doc)
		}
		return docs, nil
	}

	for i := uint(0); i < rootNode.ChildCount(); i++ {
		child := rootNode.Child(i)
		if child.Kind() != "document" {
			continue
		}
		doc, err := buildSingleDocument(child, content, filename, int(i))
		if err != nil {
			return nil, err
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// buildSingleDocument builds a Document from a tree-sitter "document" node.
func buildSingleDocument(docNode *tree_sitter.Node, content []byte, filename string, index int) (*Document, error) {
	// Find the content node (skip "---" separator).
	var contentNode *tree_sitter.Node
	for j := uint(0); j < docNode.ChildCount(); j++ {
		child := docNode.Child(j)
		if child.Kind() != "---" {
			contentNode = child
			break
		}
	}
	// For document nodes with only "---" and no content, skip.
	if contentNode == nil {
		if docNode.ChildCount() == 0 {
			return nil, nil
		}
		// Single-child document (no ---), use it directly.
		contentNode = docNode
	}

	root, err := buildAST(contentNode, content, "")
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, nil
	}

	// Extract common Tekton fields for quick access.
	var apiVersion, kind string
	if v := root.Get("apiVersion"); v != nil {
		apiVersion = v.AsScalar()
	}
	if v := root.Get("kind"); v != nil {
		kind = v.AsScalar()
	}

	return &Document{
		Filename:   filename,
		Root:       root,
		APIVersion: apiVersion,
		Kind:       kind,
		Index:      index,
	}, nil
}

// buildAST converts a tree-sitter node into our AST representation.
func buildAST(tsNode *tree_sitter.Node, content []byte, key string) (*Node, error) {
	r := nodeRange(tsNode)
	kind := tsNode.Kind()

	switch kind {
	case "stream":
		// For backward compat when called directly, take first child.
		if tsNode.ChildCount() > 0 {
			return buildAST(tsNode.Child(0), content, key)
		}
		return &Node{Key: key, Kind: NodeKindNull, Range: r}, nil

	case "document":
		// Skip "---" separator, recurse into content.
		for i := uint(0); i < tsNode.ChildCount(); i++ {
			child := tsNode.Child(i)
			if child.Kind() != "---" {
				return buildAST(child, content, key)
			}
		}
		return &Node{Key: key, Kind: NodeKindNull, Range: r}, nil

	case "block_node":
		// Wrapper — recurse into first child.
		if tsNode.ChildCount() > 0 {
			return buildAST(tsNode.Child(0), content, key)
		}
		return &Node{Key: key, Kind: NodeKindNull, Range: r}, nil

	case "block_mapping", "flow_mapping":
		mapping := make(map[string]*Node)
		for i := uint(0); i < tsNode.ChildCount(); i++ {
			child := tsNode.Child(i)
			if child.Kind() == "block_mapping_pair" || child.Kind() == "flow_pair" {
				keyNode := child.ChildByFieldName("key")
				valueNode := child.ChildByFieldName("value")
				if keyNode != nil {
					keyText := extractText(keyNode, content)
					pairRange := nodeRange(child)
					if valueNode != nil {
						valueAST, err := buildAST(valueNode, content, keyText)
						if err != nil {
							return nil, err
						}
						// Use pair range so hover/goto-definition works on the key.
						valueAST.Range = pairRange
						valueAST.Key = keyText
						mapping[keyText] = valueAST
					} else {
						mapping[keyText] = &Node{
							Key:   keyText,
							Kind:  NodeKindNull,
							Range: pairRange,
						}
					}
				}
			}
		}
		return &Node{Key: key, Kind: NodeKindMapping, MappingChildren: mapping, Range: r}, nil

	case "block_sequence", "flow_sequence":
		var items []*Node
		for i := uint(0); i < tsNode.ChildCount(); i++ {
			child := tsNode.Child(i)
			if child.Kind() == "block_sequence_item" {
				// Skip the '-' marker (child 0), take value (child 1).
				if child.ChildCount() > 1 {
					item, err := buildAST(child.Child(1), content, "")
					if err != nil {
						return nil, err
					}
					items = append(items, item)
				}
			} else if child.Kind() == "flow_node" {
				item, err := buildAST(child, content, "")
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
		}
		return &Node{Key: key, Kind: NodeKindSequence, SequenceChildren: items, Range: r}, nil

	case "plain_scalar", "single_quote_scalar", "double_quote_scalar", "block_scalar":
		text := extractText(tsNode, content)
		return &Node{Key: key, Kind: NodeKindScalar, ScalarValue: text, Range: r}, nil

	case "flow_node":
		if tsNode.ChildCount() > 0 {
			return buildAST(tsNode.Child(0), content, key)
		}
		return &Node{Key: key, Kind: NodeKindNull, Range: r}, nil

	default:
		// Try to recurse or extract text.
		if tsNode.ChildCount() > 0 {
			return buildAST(tsNode.Child(0), content, key)
		}
		text := extractText(tsNode, content)
		if text == "" {
			return &Node{Key: key, Kind: NodeKindNull, Range: r}, nil
		}
		return &Node{Key: key, Kind: NodeKindScalar, ScalarValue: text, Range: r}, nil
	}
}

// nodeRange converts a tree-sitter node position to our Range type.
func nodeRange(tsNode *tree_sitter.Node) Range {
	start := tsNode.StartPosition()
	end := tsNode.EndPosition()
	return Range{
		Start: Position{Line: uint32(start.Row), Character: uint32(start.Column)},
		End:   Position{Line: uint32(end.Row), Character: uint32(end.Column)},
	}
}

// extractText gets the text content of a tree-sitter node.
func extractText(tsNode *tree_sitter.Node, content []byte) string {
	startByte := tsNode.StartByte()
	endByte := tsNode.EndByte()
	if startByte >= uint(len(content)) || endByte > uint(len(content)) {
		return ""
	}
	return string(content[startByte:endByte])
}
