package xmltree

import (
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
)

// Convert parses XML into an ordered tree without applying EDMX compaction rules.
func Convert(content string) (Document, error) {
	content = strings.TrimPrefix(content, "\uFEFF")
	hasDeclaration, err := validateDocument(content)
	if err != nil {
		return nil, fmt.Errorf("validate XML: %w", err)
	}

	// xmlquery loses leading comments and text without a declaration. Add a
	// temporary declaration for parsing, then omit it from the converted tree.
	if !hasDeclaration {
		content = `<?xml version="1.0"?>` + content
	}
	parsed, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse XML: %w", err)
	}

	nodes, err := convertChildren(parsed)
	if err != nil {
		return nil, err
	}
	if !hasDeclaration {
		if len(nodes) == 0 || nodes[0].Kind() != "declaration" {
			return nil, fmt.Errorf("XML parser omitted the temporary declaration")
		}
		nodes = nodes[1:]
	}

	rootCount := 0

	for _, node := range nodes {
		switch node.Kind() {
		case "element":
			rootCount++
		case "text":
			if strings.TrimSpace(node.Value()) != "" {
				return nil, fmt.Errorf("XML contains text outside the document element")
			}
		case "cdata":
			return nil, fmt.Errorf("XML contains CDATA outside the document element")
		}
	}
	if rootCount != 1 {
		return nil, fmt.Errorf("XML must contain exactly one document element; found %d", rootCount)
	}

	return Document{"nodes": nodes}, nil
}

func convertChildren(parent *xmlquery.Node) ([]Node, error) {
	nodes := make([]Node, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		node, err := convertNode(child)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func convertNode(source *xmlquery.Node) (Node, error) {
	switch source.Type {
	case xmlquery.ElementNode:
		children, err := convertChildren(source)
		if err != nil {
			return nil, err
		}
		return namedNode("element", source, children), nil
	case xmlquery.DeclarationNode:
		children, err := convertChildren(source)
		if err != nil {
			return nil, err
		}
		return namedNode("declaration", source, children), nil
	case xmlquery.TextNode:
		return valueNode("text", source.Data), nil
	case xmlquery.CharDataNode:
		return valueNode("cdata", source.Data), nil
	case xmlquery.CommentNode:
		return valueNode("comment", source.Data), nil
	case xmlquery.NotationNode:
		return valueNode("directive", source.Data), nil
	case xmlquery.ProcessingInstruction:
		if source.ProcInst == nil {
			return nil, fmt.Errorf("XML processing instruction has no parsed data")
		}
		return Node{
			"kind":  "processing-instruction",
			"name":  source.ProcInst.Target,
			"value": source.ProcInst.Inst,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported XML node type: %d", source.Type)
	}
}

func namedNode(kind string, source *xmlquery.Node, children []Node) Node {
	attributes, namespaces := convertAttributes(source.Attr)

	return Node{
		"kind":                kind,
		"name":                qualifiedName(source.Prefix, source.Data),
		"localName":           source.Data,
		"prefix":              source.Prefix,
		"namespaceURI":        source.NamespaceURI,
		"attributes":          attributes,
		"attributeNamespaces": namespaces,
		"nodes":               children,
	}
}

func valueNode(kind string, value string) Node {
	return Node{
		"kind":  kind,
		"value": value,
	}
}

func convertAttributes(source []xmlquery.Attr) (map[string]string, map[string]string) {
	attributes := make(map[string]string, len(source))
	namespaces := make(map[string]string, len(source))

	for _, attribute := range source {
		name := qualifiedName(attribute.Name.Space, attribute.Name.Local)

		attributes[name] = attribute.Value
		namespaces[name] = attribute.NamespaceURI
	}
	return attributes, namespaces
}

func qualifiedName(prefix string, localName string) string {
	if prefix == "" {
		return localName
	}
	return prefix + ":" + localName
}
