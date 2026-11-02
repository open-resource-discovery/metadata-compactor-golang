//go:build unit

package xmltree

import (
	"encoding/xml"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestToXMLRoundTrip(t *testing.T) {
	cases := map[string]string{
		"literal whitespace outside root":           " \t\r\n<root/> \t\r\n",
		"edmx":                                      `<?xml version="1.0" encoding="UTF-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices><Schema Namespace="CatalogService" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="Books"><Property Name="title" Type="Edm.String"><Annotation Term="Core.Description" String="Book &amp; title"/></Property></EntityType></Schema></edmx:DataServices></edmx:Edmx>`,
		"mixed content and whitespace":              `<root xml:space="default"> before <child/> after <String>  value  </String><String> </String></root>`,
		"escaped attribute values":                  `<root value="&quot;&apos;&amp;&lt;&gt;&#9;&#10;&#13;"/>`,
		"escaped text":                              `<root>&lt;&amp;&gt;&quot;&apos;&#9;&#10;&#13;</root>`,
		"comments PI and CDATA":                     `<!--before--><?style data?><root><!--inside--><![CDATA[a < b & c]]><?target data?></root><!--after-->`,
		"PR4 ordered mixed content":                 `<?xml version="1.0"?><root id="1">hello<![CDATA[a < b]]><?run now?><child/></root>`,
		"PR4 namespace and doctype":                 `<?xml version="1.0"?><!DOCTYPE ns:root><ns:root xmlns:ns="urn:test" ns:key="a &amp; b"/>`,
		"PI without body":                           `<?run?><root/>`,
		"shared namespace prefixes":                 `<a:root xmlns:a="urn:same" xmlns:b="urn:same"><b:child/></a:root>`,
		"namespaced attribute in default namespace": `<root xmlns="urn:same" xmlns:p="urn:same" p:label="value"/>`,
		"nested prefix scope":                       `<p:root xmlns:p="urn:outer"><p:child xmlns:p="urn:inner"/><p:sibling/></p:root>`,
		"reset default namespace":                   `<root xmlns="urn:outer"><child xmlns=""/><sibling/></root>`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			document, err := Convert(input)
			if err != nil {
				t.Fatal(err)
			}
			before := semanticXML(t, input)
			output, err := document.ToXML()
			if err != nil {
				t.Fatal(err)
			}
			if after := semanticXML(t, output); !reflect.DeepEqual(before, after) {
				t.Fatalf("XML content changed\ninput: %s\noutput: %s\nbefore: %#v\nafter: %#v", input, output, before, after)
			}
			if _, err := Convert(output); err != nil {
				t.Fatalf("converter cannot read serialized XML: %v", err)
			}
		})
	}
}

// PR #4 sorts attributes and retains the spelling of existing XML prefixes.
func TestToXMLPreservesNamesAndSortsAttributes(t *testing.T) {
	document, err := Convert(`<ns:root xmlns:ns="urn:test" z="last" ns:key="a &amp; b" a="first"><ns:child/></ns:root>`)
	if err != nil {
		t.Fatal(err)
	}
	output, err := document.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	want := `<ns:root a="first" ns:key="a &amp; b" xmlns:ns="urn:test" z="last"><ns:child></ns:child></ns:root>`
	// Self-closing tags and explicit end tags represent the same empty element.
	compact := strings.ReplaceAll(output, `<ns:child/>`, `<ns:child></ns:child>`)
	if compact != want {
		t.Fatalf("qualified names or attribute ordering changed: %s", output)
	}
}

func TestToXMLPreservesDeclarationPresence(t *testing.T) {
	for _, declaration := range []string{"", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`} {
		t.Run(declaration, func(t *testing.T) {
			document, err := Convert(declaration + "<root/>")
			if err != nil {
				t.Fatal(err)
			}
			output, err := document.ToXML()
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(output, "<?xml ") != (declaration != "") {
				t.Fatalf("declaration presence changed: %s", output)
			}
			if declaration != "" && (!strings.Contains(output, `encoding="UTF-8"`) || !strings.Contains(output, `standalone="yes"`)) {
				t.Fatalf("declaration attributes changed: %s", output)
			}
		})
	}
}

func TestToXMLReflectsTreeEdits(t *testing.T) {
	document, err := Convert(`<root><keep id="first"/><remove/><keep id="last"/></root>`)
	if err != nil {
		t.Fatal(err)
	}
	root := onlyElement(t, document.Nodes())
	children := root.Nodes()
	root["nodes"] = []Node{children[0], children[2]}
	output, err := document.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	want := semanticXML(t, `<root><keep id="first"/><keep id="last"/></root>`)
	if got := semanticXML(t, output); !reflect.DeepEqual(got, want) {
		t.Fatalf("edited tree was not serialized correctly: %s", output)
	}
}

func TestToXMLIsDeterministicAndDoesNotMutateTree(t *testing.T) {
	document, err := Convert(`<root z="last" a="first" xmlns:p="urn:test" p:label="value"><child/></root>`)
	if err != nil {
		t.Fatal(err)
	}
	before, err := Convert(`<root z="last" a="first" xmlns:p="urn:test" p:label="value"><child/></root>`)
	if err != nil {
		t.Fatal(err)
	}
	first, err := document.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	second, err := document.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("serialization output is not deterministic")
	}
	if !reflect.DeepEqual(before, document) {
		t.Fatal("serialization mutated the tree")
	}
}

func TestToXMLHandlesCDATAClosingSequence(t *testing.T) {
	document, err := Convert(`<root><![CDATA[value]]></root>`)
	if err != nil {
		t.Fatal(err)
	}
	root := onlyElement(t, document.Nodes())
	root.Nodes()[0]["value"] = "before]]>after"
	output, err := document.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	want := semanticXML(t, `<root>before]]&gt;after</root>`)
	if got := semanticXML(t, output); !reflect.DeepEqual(got, want) {
		t.Fatalf("CDATA content changed: %s", output)
	}
}

func TestToXMLRejectsInvalidTrees(t *testing.T) {
	cases := map[string]Document{
		"empty":                           nil,
		"unknown kind":                    {"nodes": []Node{{"kind": "unknown"}}},
		"invalid comment":                 {"nodes": []Node{{"kind": "comment", "value": "bad--comment"}, {"kind": "element", "name": "root"}}},
		"multiple roots":                  {"nodes": []Node{{"kind": "element", "name": "first"}, {"kind": "element", "name": "second"}}},
		"invalid name":                    {"nodes": []Node{{"kind": "element", "name": "bad name"}}},
		"unbound prefix":                  {"nodes": []Node{{"kind": "element", "name": "p:root"}}},
		"invalid PI":                      {"nodes": []Node{{"kind": "processing-instruction", "name": "target", "value": "bad?>data"}, {"kind": "element", "name": "root"}}},
		"Unicode whitespace outside root": {"nodes": []Node{{"kind": "text", "value": "\u00A0"}, {"kind": "element", "name": "root"}}},
		"empty CDATA outside root":        {"nodes": []Node{{"kind": "element", "name": "root"}, {"kind": "cdata", "value": ""}}},
		"whitespace CDATA outside root":   {"nodes": []Node{{"kind": "element", "name": "root"}, {"kind": "cdata", "value": " "}}},
	}
	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			if output, err := document.ToXML(); err == nil || output != "" {
				t.Fatalf("expected error and no partial output; got %q, %v", output, err)
			}
		})
	}
}

type semanticToken struct {
	Kind  string
	Name  xml.Name
	Attrs []xml.Attr
	Value string
}

// semanticXML uses the standard XML decoder, independently of our converter.
func semanticXML(t *testing.T, content string) []semanticToken {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(content))
	var result []semanticToken
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
			var attributes []xml.Attr
			for _, attribute := range token.Attr {
				if attribute.Name.Space == "xmlns" || attribute.Name == (xml.Name{Local: "xmlns"}) {
					continue
				}
				attributes = append(attributes, attribute)
			}
			sort.Slice(attributes, func(i, j int) bool {
				if attributes[i].Name.Space != attributes[j].Name.Space {
					return attributes[i].Name.Space < attributes[j].Name.Space
				}
				return attributes[i].Name.Local < attributes[j].Name.Local
			})
			result = append(result, semanticToken{Kind: "start", Name: token.Name, Attrs: attributes})
		case xml.EndElement:
			result = append(result, semanticToken{Kind: "end", Name: token.Name})
		case xml.CharData:
			if len(result) > 0 && result[len(result)-1].Kind == "text" {
				result[len(result)-1].Value += string(token)
			} else {
				result = append(result, semanticToken{Kind: "text", Value: string(token)})
			}
		case xml.Comment:
			result = append(result, semanticToken{Kind: "comment", Value: string(token)})
		case xml.ProcInst:
			if token.Target != "xml" {
				result = append(result, semanticToken{Kind: "PI", Name: xml.Name{Local: token.Target}, Value: string(token.Inst)})
			}
		case xml.Directive:
			result = append(result, semanticToken{Kind: "directive", Value: string(token)})
		}
	}
}
