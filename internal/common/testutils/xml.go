//go:build unit || integration

package testutils

import (
	"encoding/xml"
	"io"
	"sort"
	"strings"
	"testing"
)

// AssertXMLEquals compares decoded XML, allowing escaping and empty-tag differences.
func AssertXMLEquals(t *testing.T, expected, found string) {
	t.Helper()
	AssertDeepEquals(t, xmlContent(t, expected), xmlContent(t, found))
}

type xmlContentToken struct {
	Kind  string
	Name  xml.Name
	Attrs []xml.Attr
	Value string
}

func xmlContent(t *testing.T, content string) []xmlContentToken {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(content))
	var result []xmlContentToken
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return result
		}
		if err != nil {
			t.Fatalf("invalid XML: %v\n%s", err, content)
		}
		switch token := token.(type) {
		case xml.StartElement:
			attrs := append([]xml.Attr(nil), token.Attr...)
			sort.Slice(attrs, func(i, j int) bool {
				if attrs[i].Name.Space != attrs[j].Name.Space {
					return attrs[i].Name.Space < attrs[j].Name.Space
				}
				return attrs[i].Name.Local < attrs[j].Name.Local
			})
			result = append(result, xmlContentToken{Kind: "start", Name: token.Name, Attrs: attrs})
		case xml.EndElement:
			result = append(result, xmlContentToken{Kind: "end", Name: token.Name})
		case xml.CharData:
			if len(result) > 0 && result[len(result)-1].Kind == "text" {
				result[len(result)-1].Value += string(token)
			} else {
				result = append(result, xmlContentToken{Kind: "text", Value: string(token)})
			}
		case xml.Comment:
			result = append(result, xmlContentToken{Kind: "comment", Value: string(token)})
		case xml.ProcInst:
			result = append(result, xmlContentToken{Kind: "PI", Name: xml.Name{Local: token.Target}, Value: string(token.Inst)})
		case xml.Directive:
			result = append(result, xmlContentToken{Kind: "directive", Value: string(token)})
		}
	}
}
