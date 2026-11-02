package xmltree

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// ToXML serializes the tree without applying compaction rules.
func (d Document) ToXML() (string, error) {
	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	bindings := map[string]string{"xml": xmlNamespaceURI}

	for _, node := range d.Nodes() {
		if node.Kind() == "text" {
			if strings.Trim(node.Value(), " \t\r\n") != "" {
				return "", fmt.Errorf("only XML whitespace is allowed outside the document element")
			}
			// Escaped whitespace becomes an entity, which is invalid outside the root.
			if err := encoder.Flush(); err != nil {
				return "", fmt.Errorf("flush XML: %w", err)
			}
			output.WriteString(node.Value())
			continue
		}
		if err := writeNode(encoder, &output, node, bindings); err != nil {
			return "", fmt.Errorf("serialize XML: %w", err)
		}
	}
	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("flush XML: %w", err)
	}
	result := output.String()
	if _, err := validateDocument(result); err != nil {
		return "", fmt.Errorf("validate serialized XML: %w", err)
	}
	return result, nil
}

func writeNode(encoder *xml.Encoder, output *bytes.Buffer, node Node, bindings map[string]string) error {
	switch node.Kind() {
	case "element":
		start, childBindings, err := elementToken(node, bindings)
		if err != nil {
			return err
		}
		if err := encoder.EncodeToken(start); err != nil {
			return err
		}
		for _, child := range node.Nodes() {
			if err := writeNode(encoder, output, child, childBindings); err != nil {
				return err
			}
		}
		return encoder.EncodeToken(start.End())

	case "text":
		return encoder.EncodeToken(xml.CharData(node.Value()))

	case "cdata":
		if err := encoder.Flush(); err != nil {
			return err
		}
		value := strings.ReplaceAll(node.Value(), "]]>", "]]]]><![CDATA[>")
		output.WriteString("<![CDATA[" + value + "]]>")
		return nil

	case "comment":
		if strings.Contains(node.Value(), "--") || strings.HasSuffix(node.Value(), "-") {
			return fmt.Errorf("invalid XML comment")
		}
		return encoder.EncodeToken(xml.Comment(node.Value()))

	case "processing-instruction":
		if strings.EqualFold(node.Name(), "xml") {
			return fmt.Errorf("use a declaration node for the XML declaration")
		}
		return encoder.EncodeToken(xml.ProcInst{
			Target: node.Name(),
			Inst:   []byte(node.Value()),
		})

	case "declaration":
		if node.Name() != "xml" {
			return fmt.Errorf("unsupported declaration %q", node.Name())
		}
		if version, found := node.Attribute("version"); !found || version != "1.0" {
			return fmt.Errorf("XML declaration must specify version 1.0")
		}
		var instruction bytes.Buffer
		for _, name := range []string{"version", "encoding", "standalone"} {
			if value, found := node.Attribute(name); found {
				if instruction.Len() != 0 {
					instruction.WriteByte(' ')
				}
				instruction.WriteString(name + "=\"")
				if err := xml.EscapeText(&instruction, []byte(value)); err != nil {
					return err
				}
				instruction.WriteByte('"')
			}
		}
		return encoder.EncodeToken(xml.ProcInst{Target: "xml", Inst: instruction.Bytes()})

	case "directive":
		return encoder.EncodeToken(xml.Directive(node.Value()))

	default:
		return fmt.Errorf("unsupported node kind %q", node.Kind())
	}
}

func elementToken(node Node, inherited map[string]string) (xml.StartElement, map[string]string, error) {
	bindings := make(map[string]string, len(inherited))
	for prefix, uri := range inherited {
		bindings[prefix] = uri
	}
	attributes := make(map[string]string, len(node.Attributes()))
	for name, value := range node.Attributes() {
		attributes[name] = value
		if name == "xmlns" {
			bindings[""] = value
		} else if strings.HasPrefix(name, "xmlns:") {
			bindings[strings.TrimPrefix(name, "xmlns:")] = value
		}
	}

	bindName := func(name, uri string, allowDefault bool) (string, error) {
		prefix, local := splitName(name)
		if local == "" {
			return "", fmt.Errorf("empty XML name")
		}
		if uri == "" {
			if prefix != "" {
				return "", fmt.Errorf("name %q has no namespace URI", name)
			}
			if allowDefault && bindings[""] != "" {
				attributes["xmlns"] = ""
				bindings[""] = ""
			}
			return local, nil
		}
		if bindings[prefix] == uri && (allowDefault || prefix != "") {
			return qualifiedName(prefix, local), nil
		}
		for _, candidate := range sortedKeys(bindings) {
			if bindings[candidate] == uri && (allowDefault || candidate != "") {
				return qualifiedName(candidate, local), nil
			}
		}
		for index := 1; ; index++ {
			candidate := fmt.Sprintf("ns%d", index)
			if _, exists := bindings[candidate]; !exists {
				bindings[candidate] = uri
				attributes["xmlns:"+candidate] = uri
				return qualifiedName(candidate, local), nil
			}
		}
	}

	name, err := bindName(node.Name(), node.NamespaceURI(), true)
	if err != nil {
		return xml.StartElement{}, nil, err
	}
	for _, original := range sortedKeys(node.Attributes()) {
		if original == "xmlns" || strings.HasPrefix(original, "xmlns:") {
			continue
		}
		name, err := bindName(original, node.AttributeNamespace(original), false)
		if err != nil {
			return xml.StartElement{}, nil, err
		}
		value := attributes[original]
		delete(attributes, original)
		attributes[name] = value
	}

	start := xml.StartElement{Name: xml.Name{Local: name}}
	for _, name := range sortedKeys(attributes) {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: name},
			Value: attributes[name],
		})
	}
	return start, bindings, nil
}

func splitName(name string) (string, string) {
	prefix, local, found := strings.Cut(name, ":")
	if !found {
		return "", name
	}
	return prefix, local
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
