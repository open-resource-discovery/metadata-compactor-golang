//go:build unit

package xmltree

import "testing"

func TestConvertPreservesElementsNamespacesAndValues(t *testing.T) {
	for _, content := range []string{
		`<root xmlns="urn:test" xmlns:m="urn:attributes" m:label="A &amp; B" empty=""><child Name="first">  value  </child><child Name="second"/></root>`,
		`<t:root xmlns:t="urn:test" xmlns:m="urn:attributes" m:label="A &amp; B" empty=""><t:child Name="first">  value  </t:child><t:child Name="second"/></t:root>`,
	} {
		t.Run(content, func(t *testing.T) {
			document, err := Convert(content)
			if err != nil {
				t.Fatal(err)
			}
			root := onlyElement(t, document.Nodes())
			if root.LocalName() != "root" || root.NamespaceURI() != "urn:test" {
				t.Fatalf("unexpected root identity: %q in %q", root.LocalName(), root.NamespaceURI())
			}
			if value, found := root.Attribute("m:label"); !found || value != "A & B" {
				t.Fatalf("escaped attribute = %q, found = %v", value, found)
			}
			if got := root.AttributeNamespace("m:label"); got != "urn:attributes" {
				t.Fatalf("attribute namespace = %q", got)
			}
			if value, found := root.Attribute("empty"); !found || value != "" {
				t.Fatal("empty attribute was lost")
			}
			if _, found := root.Attribute("missing"); found {
				t.Fatal("missing attribute reported as present")
			}
			children := root.Nodes()
			if len(children) != 2 {
				t.Fatalf("got %d children; want 2", len(children))
			}
			for index, name := range []string{"first", "second"} {
				if got, _ := children[index].Attribute("Name"); got != name {
					t.Fatalf("child %d name = %q; want %q", index, got, name)
				}
				if children[index].NamespaceURI() != "urn:test" {
					t.Fatal("child namespace was lost")
				}
			}
			text := children[0].Nodes()
			if len(text) != 1 || text[0].Value() != "  value  " {
				t.Fatalf("text whitespace changed: %#v", text)
			}
		})
	}
}

func TestConvertPreservesNonElementChildren(t *testing.T) {
	document, err := Convert(`<?xml version="1.0"?><root> <!--note--><![CDATA[a < b & c]]><?target data?></root>`)
	if err != nil {
		t.Fatal(err)
	}
	children := onlyElement(t, document.Nodes()).Nodes()
	if len(children) != 4 {
		t.Fatalf("got %d children; want 4", len(children))
	}
	for index, kind := range []string{"text", "comment", "cdata", "processing-instruction"} {
		if children[index].Kind() != kind {
			t.Fatalf("child %d kind = %q; want %q", index, children[index].Kind(), kind)
		}
	}
	if children[0].Value() != " " || children[1].Value() != "note" || children[2].Value() != "a < b & c" {
		t.Fatal("text, comment, or CDATA value changed")
	}
	if children[3].Name() != "target" || children[3].Value() != "data" {
		t.Fatal("processing instruction changed")
	}
}

func TestConvertRejectsInvalidDocuments(t *testing.T) {
	for _, content := range []string{
		"", " \n", "<root>", "<root></other>", "<first/><second/>",
		"text<root/>", "<root/>text", "<root/><![CDATA[value]]>",
		"\u00A0<root/>", "<root/>\u2003", "&#32;<root/>",
		"<root/><![CDATA[]]>", "<root/><![CDATA[ ]]>",
	} {
		t.Run(content, func(t *testing.T) {
			if _, err := Convert(content); err == nil {
				t.Fatal("expected invalid XML to be rejected")
			}
		})
	}
}

func TestConvertPreservesDocumentProlog(t *testing.T) {
	for _, declaration := range []string{"", `<?xml version="1.0"?>`} {
		t.Run(declaration, func(t *testing.T) {
			document, err := Convert(declaration + `<!--before--><?style data?><root/><!--after-->`)
			if err != nil {
				t.Fatal(err)
			}
			nodes := document.Nodes()
			if declaration != "" {
				if nodes[0].Kind() != "declaration" {
					t.Fatal("explicit declaration was lost")
				}
				nodes = nodes[1:]
			}
			if len(nodes) != 4 {
				t.Fatalf("got %d nodes; want 4", len(nodes))
			}
			for index, kind := range []string{"comment", "processing-instruction", "element", "comment"} {
				if nodes[index].Kind() != kind {
					t.Fatalf("node %d kind = %q; want %q", index, nodes[index].Kind(), kind)
				}
			}
			if nodes[0].Value() != "before" || nodes[3].Value() != "after" {
				t.Fatal("document comment values changed")
			}
		})
	}
}

func TestConvertAcceptsUTF8BOM(t *testing.T) {
	document, err := Convert("\uFEFF<root/>")
	if err != nil {
		t.Fatal(err)
	}
	if onlyElement(t, document.Nodes()).LocalName() != "root" {
		t.Fatal("unexpected document element")
	}
}

func onlyElement(t *testing.T, nodes []Node) Node {
	t.Helper()
	var result Node
	for _, node := range nodes {
		if node.Kind() == "element" {
			if result != nil {
				t.Fatal("more than one document element")
			}
			result = node
		}
	}
	if result == nil {
		t.Fatal("document element not found")
	}
	return result
}

func TestConvertRejectsDuplicateAttributes(t *testing.T) {
	tests := []struct{ name, content string }{
		{"unqualified attributes", `<Schema Namespace="First" Namespace="Second"/>`},
		{"qualified attributes", `<root xmlns:p="urn:test" p:key="one" p:key="two"/>`},
		{"same expanded name", `<root xmlns:a="urn:test" xmlns:b="urn:test" a:key="one" b:key="two"/>`},
		{"duplicate namespace declarations", `<root xmlns:p="urn:first" xmlns:p="urn:second"/>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if document, err := Convert(tc.content); err == nil || document != nil {
				t.Fatalf("expected duplicate-attribute error without a tree; got %#v, %v", document, err)
			}
		})
	}
}

func TestConvertRejectsReservedNamespaceBindings(t *testing.T) {
	tests := []struct{ name, content string }{
		{"xml prefix rebound", `<root xmlns:xml="urn:wrong"/>`},
		{"xml prefix cleared", `<root xmlns:xml=""/>`},
		{"xml namespace aliased", `<root xmlns:p="http://www.w3.org/XML/1998/namespace"/>`},
		{"xml namespace made default", `<root xmlns="http://www.w3.org/XML/1998/namespace"/>`},
		{"xmlns prefix declared", `<root xmlns:xmlns="urn:wrong"/>`},
		{"xmlns namespace aliased", `<root xmlns:p="http://www.w3.org/2000/xmlns/"/>`},
		{"xmlns namespace made default", `<root xmlns="http://www.w3.org/2000/xmlns/"/>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if document, err := Convert(tc.content); err == nil || document != nil {
				t.Fatalf("expected namespace error without a tree; got %#v, %v", document, err)
			}
		})
	}
}

func TestConvertAcceptsDistinctAttributesAndValidNamespaceBindings(t *testing.T) {
	tests := []struct{ name, content string }{
		{"different attribute namespaces", `<root a:key="one" b:key="two" xmlns:a="urn:first" xmlns:b="urn:second"/>`},
		{"default namespace does not apply to attributes", `<root xmlns="urn:test" xmlns:p="urn:test" key="one" p:key="two"/>`},
		{"shared namespace with distinct names", `<root xmlns:a="urn:test" xmlns:b="urn:test" a:first="one" b:second="two"/>`},
		{"explicit xml binding", `<root xmlns:xml="http://www.w3.org/XML/1998/namespace" xml:lang="en"/>`},
		{"implicit xml binding", `<root xml:lang="en"/>`},
		{"namespace URI named xmlns", `<root xmlns:p="xmlns" p:xml="ordinary" p:p="value"/>`},
		{"nested namespace scopes", `<root xmlns:a="urn:first" xmlns:b="urn:first"><child xmlns:b="urn:second" a:key="one" b:key="two"/><sibling xmlns:b="urn:third" a:key="one" b:key="two"/></root>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			document, err := Convert(tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := document.ToXML(); err != nil {
				t.Fatalf("valid document cannot be serialized: %v", err)
			}
		})
	}
}
