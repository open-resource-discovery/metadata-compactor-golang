package xmltree

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	xmlNamespaceURI   = "http://www.w3.org/XML/1998/namespace"
	xmlnsNamespaceURI = "http://www.w3.org/2000/xmlns/"
)

// validateDocument checks the original input before xmlquery can discard nodes.
func validateDocument(content string) (bool, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	// Token expands namespace prefixes. Keep raw names alongside it so namespace
	// declarations can be distinguished from ordinary namespaced attributes.
	rawDecoder := xml.NewDecoder(strings.NewReader(content))
	depth, roots, tokenCount := 0, 0, 0
	hasDeclaration := false

	for {
		startOffset := decoder.InputOffset()
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		rawToken, err := rawDecoder.RawToken()
		if err != nil {
			return false, err
		}
		switch token := token.(type) {
		case xml.StartElement:
			rawElement, ok := rawToken.(xml.StartElement)
			if !ok {
				return false, fmt.Errorf("inconsistent XML start element")
			}
			if err := validateAttributes(token, rawElement); err != nil {
				return false, err
			}
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			// Outside the root, XML permits only literal spaces, tabs, CR, and LF.
			// Check the source so entities and CDATA cannot masquerade as whitespace.
			if depth == 0 && strings.Trim(content[startOffset:decoder.InputOffset()], " \t\r\n") != "" {
				return false, fmt.Errorf("only XML whitespace is allowed outside the document element")
			}
		case xml.ProcInst:
			if token.Target == "xml" {
				if tokenCount != 0 {
					return false, fmt.Errorf("XML declaration must appear first")
				}
				hasDeclaration = true
			}
		}
		tokenCount++
	}
	if roots != 1 {
		return false, fmt.Errorf("expected one document element; found %d", roots)
	}
	return hasDeclaration, nil
}

func validateAttributes(element, rawElement xml.StartElement) error {
	seen := make(map[xml.Name]struct{}, len(element.Attr))

	for index, attribute := range element.Attr {
		rawName := rawElement.Attr[index].Name
		name := attribute.Name
		if rawName.Space == "xmlns" || rawName == (xml.Name{Local: "xmlns"}) {
			prefix := ""
			if rawName.Space == "xmlns" {
				prefix = rawName.Local
			}
			uri := attribute.Value
			switch {
			case prefix == "xmlns":
				return fmt.Errorf("the xmlns prefix must not be declared")
			case prefix == "xml" && uri != xmlNamespaceURI:
				return fmt.Errorf("the xml prefix must be bound to %q", xmlNamespaceURI)
			case prefix != "xml" && uri == xmlNamespaceURI:
				return fmt.Errorf("only the xml prefix may bind namespace %q", xmlNamespaceURI)
			case uri == xmlnsNamespaceURI:
				return fmt.Errorf("namespace %q must not be declared", xmlnsNamespaceURI)
			}
			name = xml.Name{Space: xmlnsNamespaceURI, Local: prefix}
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate XML attribute %q in namespace %q", name.Local, name.Space)
		}
		seen[name] = struct{}{}
	}
	return nil
}
