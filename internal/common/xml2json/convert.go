// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and overlay-golang contributors
// SPDX-License-Identifier: Apache-2.0
// Adapted from overlay-golang/internal/common/xml2json for metadata compaction.

package xml2json

import (
	"strings"
	"unicode"

	"github.com/antchfx/xmlquery"
)

// IsWhitespaceOnly follows JavaScript trim: Unicode whitespace plus BOM, excluding NEL.
// The converter and EDMX target resolver share this check for Node compatibility.
func IsWhitespaceOnly(value string) bool {
	return strings.TrimFunc(value, func(r rune) bool {
		return r == '\uFEFF' || (r != '\u0085' && unicode.IsSpace(r))
	}) == ""
}

func Convert(xml string) (Document, error) {
	content := strings.TrimPrefix(xml, "\uFEFF")
	hasDeclaration := strings.HasPrefix(content, "<?xml") && len(content) > 5 && strings.ContainsRune(" \t\r\n", rune(content[5]))
	if !hasDeclaration {
		// xmlquery needs a declaration to retain a root preceded by a DOCTYPE.
		content = "<?xml version=\"1.0\"?>" + content
	}
	parsed, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	var prolog []string
	if hasDeclaration {
		prolog = declarations(parsed)
	}
	return NewDocument(nodes(parsed), notations(parsed), prolog), nil
}

func nodes(parent *xmlquery.Node) []Node {
	result := make([]Node, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case xmlquery.TextNode:
			if !IsWhitespaceOnly(child.Data) {
				result = append(result, NewTextNode(child.Data))
			}
		case xmlquery.ElementNode:
			result = append(result, NewElementNode(name(child.Prefix, child.Data), nodes(child), attributes(child.Attr...)))
		case xmlquery.CharDataNode:
			result = append(result, NewCdataNode(child.Data))
		case xmlquery.ProcessingInstruction:
			result = append(result, NewProcessingInstructionNode(child.OutputXML(true)))
		}
	}

	return result
}

func name(namespace, local string) string {
	if namespace == "" {
		return local
	}
	return namespace + ":" + local
}

func notations(parent *xmlquery.Node) []string {
	result := make([]string, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.NotationNode {
			result = append(result, child.OutputXML(true))
		}
	}

	return result
}

func declarations(parent *xmlquery.Node) []string {
	result := make([]string, 0)

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xmlquery.DeclarationNode {
			result = append(result, child.OutputXML(true))
		}
	}

	return result
}

func attributes(attributes ...xmlquery.Attr) Attributes {
	result := make(Attributes, len(attributes))
	for _, attribute := range attributes {
		result[name(attribute.Name.Space, attribute.Name.Local)] = attribute.Value
	}
	return result
}
