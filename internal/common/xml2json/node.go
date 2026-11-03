// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and overlay-golang contributors
// SPDX-License-Identifier: Apache-2.0
// Adapted from overlay-golang/internal/common/xml2json for metadata compaction.

package xml2json

import "github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"

type Node map[string]any

func NewTextNode(value string) Node {
	return map[string]any{
		"type":  "text",
		"value": value,
	}
}

func NewCdataNode(value string) Node {
	return map[string]any{
		"type":  "cdata",
		"value": value,
	}
}

func NewProcessingInstructionNode(value string) Node {
	return map[string]any{
		"type":  "processing-instruction",
		"value": value,
	}
}

func NewElementNode(name string, nodes []Node, attributes Attributes) Node {
	return map[string]any{
		"type":       "element",
		"name":       name,
		"nodes":      nodes,
		"attributes": attributes,
	}
}

func (self Node) Type() string {
	return utils.SafeCast[string](self["type"])
}

func (self Node) Name() string {
	return utils.SafeCast[string](self["name"])
}

func (self Node) Value() string {
	return utils.SafeCast[string](self["value"])
}

func (self Node) Nodes() []Node {
	return utils.SafeCast[[]Node](self["nodes"])
}

func (self Node) Attributes() Attributes {
	return utils.SafeCast[Attributes](self["attributes"])
}

func (self Node) Attribute(name string) string {
	return self.Attributes().Get(name)
}
