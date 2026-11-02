//go:build unit

package xml2json

import (
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

func TestConvertRoundTrip(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{
			name:     "namespaces and escaped attributes",
			input:    `<edm:Schema xmlns:edm="urn:edm" xmlns:m="urn:metadata" m:label="A &amp; B" empty=""><edm:String>  value  </edm:String></edm:Schema>`,
			expected: `<edm:Schema xmlns:edm="urn:edm" xmlns:m="urn:metadata" m:label="A &amp; B" empty=""><edm:String>  value  </edm:String></edm:Schema>`,
		},
		{
			name:     "Node whitespace classification",
			input:    `<root><String>&#xFEFF;</String><String>&#x85;</String><String> &#xFEFF; </String></root>`,
			expected: "<root><String/><String>\u0085</String><String/></root>",
		},
		{
			name:     "CDATA whitespace is preserved",
			input:    `<root><![CDATA[ ]]></root>`,
			expected: `<root><![CDATA[ ]]></root>`,
		},
		{
			name:     "comments and whitespace removed",
			input:    `<!--before--><root> <child/> <!--inside--> </root><!--after-->`,
			expected: `<root><child/></root>`,
		},
		{
			name:     "mixed content retained",
			input:    `<root>hello &amp; world<![CDATA[a < b]]><?run now?><child/> tail </root>`,
			expected: `<root>hello &amp; world<![CDATA[a < b]]><?run now?><child/> tail </root>`,
		},
		{
			name:     "explicit declaration and doctype",
			input:    `<?xml version="1.0"?><?run now?><!DOCTYPE root><root/>`,
			expected: `<?xml version="1.0"?><!DOCTYPE root><?run now?><root/>`,
		},
		{
			name:     "doctype without declaration retains root",
			input:    `<!DOCTYPE root><root><child/></root>`,
			expected: `<!DOCTYPE root><root><child/></root>`,
		},
		{
			name:     "UTF8 byte order mark",
			input:    "\uFEFF<root/>",
			expected: `<root/>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			document, err := Convert(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertXMLEquals(t, tc.expected, document.ToXML())
		})
	}
}

func TestConvertAttributes(t *testing.T) {
	document, err := Convert(`<root empty="" label="A &amp; B"/>`)
	if err != nil {
		t.Fatal(err)
	}
	root := document.Nodes()[0]
	if !root.Attributes().Has("empty") || root.Attributes().Has("missing") {
		t.Fatal("empty and missing attributes must remain distinguishable")
	}
	if root.Attribute("label") != "A & B" {
		t.Fatal("attribute entities were not decoded")
	}
}

func TestConvertParserErrors(t *testing.T) {
	for _, input := range []string{"", "<root>", "<root></other>"} {
		if document, err := Convert(input); err == nil || document != nil {
			t.Fatalf("expected parser error for %q; got %v, %v", input, document, err)
		}
	}
}

func TestConvertPreservesWhitespaceCharacterReferences(t *testing.T) {
	for _, input := range []string{
		`<root value="a&#x9;b&#xA;c&#xD;d"/>`,
		`<root>a&#xD;b</root>`,
	} {
		document, err := Convert(input)
		if err != nil {
			t.Fatal(err)
		}
		// Literal whitespace would be normalized by XML consumers on reparse.
		if output := document.ToXML(); output != input {
			t.Fatalf("character references changed: got %q, want %q", output, input)
		}
	}
}
