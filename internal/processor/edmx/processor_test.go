//go:build unit

package edmx

import (
	"strings"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

func TestEDMXProcessorProcess(t *testing.T) {
	tests := []struct {
		name     string
		document string
		rules    []edmx.EDMXRule
		expected string
	}{
		{
			name:     "global exact",
			document: `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="Core.Description" String="keep"/><Annotation Term="Core.LongDescription"/></EntityType></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "Core.Description"}},
			expected: `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation String="keep" Term="Core.Description"/></EntityType></Schema>`,
		},
		{
			name:     "prefix keeps dot boundary",
			document: `<Schema Namespace="Demo"><Annotation Term="Common.Label"/><Annotation Term="Common.LabelExtra"/><Annotation Term="CommonExtra.Label"/><Annotation Term="Common"/></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "glob", Value: "Common.*"}},
			expected: `<Schema Namespace="Demo"><Annotation Term="Common.Label"/><Annotation Term="Common.LabelExtra"/></Schema>`,
		},
		{
			name:     "qualifiers ignored by global rules",
			document: `<Schema Namespace="Demo"><Annotation Term="Core.Description"/><Annotation Term="Core.Description" Qualifier="Short"/></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "Core.Description"}},
			expected: `<Schema Namespace="Demo"><Annotation Term="Core.Description"/><Annotation Qualifier="Short" Term="Core.Description"/></Schema>`,
		},
		{
			name:     "nested annotations pruned",
			document: `<edm:Schema xmlns:edm="urn:edm"><edm:Annotation Term="Core.Description"><edm:Record><edm:Annotation Term="Other.Hidden"/><edm:PropertyValue Property="Label" String="keep"/></edm:Record></edm:Annotation></edm:Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "glob", Value: "Core.*"}},
			expected: `<edm:Schema xmlns:edm="urn:edm"><edm:Annotation Term="Core.Description"><edm:Record><edm:PropertyValue Property="Label" String="keep"/></edm:Record></edm:Annotation></edm:Schema>`,
		},
		{
			name:     "empty external wrappers retained",
			document: `<Schema><Annotations Target="Demo.Product"><Annotation Term="One"/><Annotation/></Annotations></Schema>`,
			rules:    []edmx.EDMXRule{},
			expected: `<Schema><Annotations Target="Demo.Product"/></Schema>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CreateProcessor(tc.rules).Process(tc.document)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertXMLEquals(t, tc.expected, got)
		})
	}
}

func TestEDMXProcessorCopiesConfiguredRules(t *testing.T) {
	target, qualifier := "Demo.Product", "Summary"
	rules := []edmx.EDMXRule{{Kind: "exact", Value: "One", Target: &target, Qualifier: &qualifier}}
	processor := CreateProcessor(rules)
	rules[0].Value = "Changed"
	target, qualifier = "Demo.Other", "Other"
	input := `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="One" Qualifier="Summary"/><Annotation Term="One" Qualifier="Other"/></EntityType><EntityType Name="Other"><Annotation Term="One" Qualifier="Other"/></EntityType></Schema>`
	expected := `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="One" Qualifier="Summary"/></EntityType><EntityType Name="Other"/></Schema>`
	output, err := processor.Process(input)
	if err != nil {
		t.Fatal(err)
	}
	testutils.AssertXMLEquals(t, expected, output)
}

func TestEDMXProcessorBaselinesDoNotLeakBetweenCalls(t *testing.T) {
	processor := CreateProcessor(nil)
	input := `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="One" String="current"/></EntityType></Schema>`
	baseline := `<Schema><Annotations Target="Demo.Product"><Annotation Term="One" String="old"/></Annotations></Schema>`
	first, err := processor.Process(input, baseline)
	if err != nil || !strings.Contains(first, `String="current"`) {
		t.Fatalf("first = %s, %v", first, err)
	}
	second, err := processor.Process(input)
	if err != nil || strings.Contains(second, "Annotation") {
		t.Fatalf("baseline leaked: %s, %v", second, err)
	}
}

func TestEDMXProcessorErrors(t *testing.T) {
	processor := CreateProcessor(nil)
	for _, c := range []struct {
		input     string
		baselines []string
	}{
		{"<unclosed>", nil}, {"", nil}, {"<root/>", []string{"<unclosed>"}},
	} {
		if output, err := processor.Process(c.input, c.baselines...); err == nil || output != "" {
			t.Fatalf("expected error without partial output, got %q, %v", output, err)
		}
	}
}

// These focused examples retain the reviewed Node PR #5 expected outputs.
func TestEDMXProcessorNodeCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		document  string
		rules     []edmx.EDMXRule
		baselines []string
		expected  string
	}{
		{
			name:     "at sign is literal",
			document: `<Schema Namespace="Demo"><Annotation Term="Core.Description"/></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "@Core.Description"}},
			expected: `<Schema Namespace="Demo"/>`,
		},
		{
			name:     "removed parent removes preserved child",
			document: `<Schema><Annotation Term="Other.Parent"><Annotation Term="Core.Description"/></Annotation></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "Core.Description"}},
			expected: `<Schema/>`,
		},
		{
			name:      "unqualified baseline rejects qualified annotations",
			document:  `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="Baseline.Term" Qualifier="kept" String="current"/></Annotations><Annotations Target="Demo.Other"><Annotation Term="Baseline.Term" Qualifier="removed"/></Annotations></Schema>`,
			baselines: []string{`<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="Baseline.Term" String="old"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><Annotations Target="Demo.Product"/><Annotations Target="Demo.Other"/></Schema>`,
		},
		{
			name:      "parameter order distinguishes operations",
			document:  `<Schema Namespace="Demo"><Function Name="Search"><Parameter Name="a" Type="Edm.String"/><Parameter Name="b" Type="Edm.Int32"/><Annotation Term="Operation.Term"/></Function><Function Name="Search"><Parameter Name="a" Type="Edm.Int32"/><Parameter Name="b" Type="Edm.String"/><Annotation Term="Operation.Term"/></Function></Schema>`,
			baselines: []string{`<Schema><Annotations Target="Demo.Search(Edm.String,Edm.Int32)"><Annotation Term="Operation.Term"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><Function Name="Search"><Parameter Name="a" Type="Edm.String"/><Parameter Name="b" Type="Edm.Int32"/><Annotation Term="Operation.Term"/></Function><Function Name="Search"><Parameter Name="a" Type="Edm.Int32"/><Parameter Name="b" Type="Edm.String"/></Function></Schema>`,
		},
		{
			name:      "literal external operation target",
			document:  `<Schema Namespace="Demo"><Function Name="Search"><Annotation Term="Operation.Term"/></Function><Annotations Target="Demo.Search"><Annotation Term="Operation.Term"/></Annotations></Schema>`,
			baselines: []string{`<Schema><Annotations Target="Demo.Search"><Annotation Term="Operation.Term"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><Function Name="Search"/><Annotations Target="Demo.Search"><Annotation Term="Operation.Term"/></Annotations></Schema>`,
		},
		{
			name:      "container member targets",
			document:  `<Schema Namespace="Demo"><EntityContainer Name="Container"><EntitySet Name="Products"><Annotation Term="One"/></EntitySet><FunctionImport Name="Search"><Annotation Term="Two"/></FunctionImport></EntityContainer><EnumType Name="Color"><Member Name="Red"><Annotation Term="Three"/></Member></EnumType></Schema>`,
			baselines: []string{`<Schema><Annotations Target="Demo.Container/Products"><Annotation Term="One"/></Annotations><Annotations Target="Demo.Container/Search"><Annotation Term="Two"/></Annotations><Annotations Target="Demo.Color/Red"><Annotation Term="Three"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><EntityContainer Name="Container"><EntitySet Name="Products"><Annotation Term="One"/></EntitySet><FunctionImport Name="Search"><Annotation Term="Two"/></FunctionImport></EntityContainer><EnumType Name="Color"><Member Name="Red"><Annotation Term="Three"/></Member></EnumType></Schema>`,
		},
		{
			name:     "local names also match foreign namespaces",
			document: `<root xmlns:x="urn:foreign"><x:Annotation Term="Remove"/><x:Annotation Term="Keep"/></root>`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "Keep"}},
			expected: `<root xmlns:x="urn:foreign"><x:Annotation Term="Keep"/></root>`,
		},
		{
			name:      "unresolved baseline ignored",
			document:  `<root><Annotation Term="Keep"/></root>`,
			baselines: []string{`<root><Annotation Term="Keep"/></root>`},
			expected:  `<root/>`,
		},
		{
			name:      "empty baseline target is ignored",
			document:  `<Schema><Annotations Target="Anywhere"><Annotation Term="Keep"/></Annotations></Schema>`,
			baselines: []string{`<Schema><Annotations Target=""><Annotation Term="Keep"/></Annotations></Schema>`},
			expected:  `<Schema><Annotations Target="Anywhere"/></Schema>`,
		},
		{
			name:     "XML mixed content and escaping",
			document: `<?xml version="1.0"?><root quote='a " b'>hello &amp; world<![CDATA[a < b]]><?run now?><child/></root>`,
			expected: `<?xml version="1.0"?><root quote="a &quot; b">hello &amp; world<![CDATA[a < b]]><?run now?><child/></root>`,
		},
		{
			name: "comments and whitespace only text discarded",
			document: `<!--before--><Schema Namespace="Demo">
  <Annotation Term="Keep"><String> </String><String>  value  </String><!--inside--></Annotation>
</Schema><!--after-->`,
			rules:    []edmx.EDMXRule{{Kind: "exact", Value: "Keep"}},
			expected: `<Schema Namespace="Demo"><Annotation Term="Keep"><String/><String>  value  </String></Annotation></Schema>`,
		},
		{
			name:     "DOCTYPE and processing instruction order",
			document: `<?xml version="1.0"?><?run now?><!DOCTYPE ns:root><ns:root xmlns:ns="urn:test" ns:key="a &amp; b"/>`,
			expected: `<?xml version="1.0"?><!DOCTYPE ns:root><?run now?><ns:root ns:key="a &amp; b" xmlns:ns="urn:test"/>`,
		},
		{
			name:     "JavaScript whitespace classification",
			document: `<Schema><String>&#xFEFF;</String><String>&#x85;</String></Schema>`,
			expected: `<Schema><String/><String>` + "\u0085" + `</String></Schema>`,
		},
		{
			name:      "block qualifier input matches inline baseline",
			document:  `<Schema Namespace="Demo"><Annotations Target="Demo.Product" Qualifier="A"><Annotation Term="V.T" String="keep"/></Annotations><Annotations Target="Demo.Product" Qualifier="B"><Annotation Term="V.T" String="remove"/></Annotations></Schema>`,
			baselines: []string{`<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T" Qualifier="A"/></EntityType></Schema>`},
			expected:  `<Schema Namespace="Demo"><Annotations Qualifier="A" Target="Demo.Product"><Annotation String="keep" Term="V.T"/></Annotations><Annotations Qualifier="B" Target="Demo.Product"/></Schema>`,
		},
		{
			name:      "inline qualifier input matches block baseline",
			document:  `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T" Qualifier="A" String="keep"/><Annotation Term="V.T" Qualifier="B" String="remove"/><Annotation Term="V.T" String="remove"/></EntityType></Schema>`,
			baselines: []string{`<Schema Namespace="Demo"><Annotations Target="Demo.Product" Qualifier="A"><Annotation Term="V.T"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Qualifier="A" String="keep" Term="V.T"/></EntityType></Schema>`,
		},
		{
			name:      "nested qualifiers ignore distant blocks",
			document:  `<Schema Namespace="Demo"><Annotations Target="Demo.Product" Qualifier="A"><Annotation Term="V.Outer"><Annotation Term="V.Inner" String="keep"/><Annotation Term="V.Inner" Qualifier="A" String="remove"/></Annotation></Annotations></Schema>`,
			baselines: []string{`<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer" Qualifier="A"><Annotation Term="V.Inner"/></Annotation></EntityType></Schema>`},
			expected:  `<Schema Namespace="Demo"><Annotations Qualifier="A" Target="Demo.Product"><Annotation Term="V.Outer"><Annotation String="keep" Term="V.Inner"/></Annotation></Annotations></Schema>`,
		},
		{
			name:      "new schema and container member targets",
			document:  `<Schema Namespace="Demo"><TypeDefinition Name="Code"><Annotation Term="V.T"/></TypeDefinition><Term Name="Label"><Annotation Term="V.T"/></Term><EntityType Name="Product"><Annotation Term="V.T"/></EntityType><EntityContainer Name="Service"><Singleton Name="Main"><Annotation Term="V.T"/></Singleton><ActionImport Name="Create"><Annotation Term="V.T"/></ActionImport></EntityContainer></Schema>`,
			baselines: []string{`<Schema><Annotations Target="Demo.Code"><Annotation Term="V.T"/></Annotations><Annotations Target="Demo.Label"><Annotation Term="V.T"/></Annotations><Annotations Target="Demo.Service/Main"><Annotation Term="V.T"/></Annotations><Annotations Target="Demo.Service/Create"><Annotation Term="V.T"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><TypeDefinition Name="Code"><Annotation Term="V.T"/></TypeDefinition><Term Name="Label"><Annotation Term="V.T"/></Term><EntityType Name="Product"/><EntityContainer Name="Service"><Singleton Name="Main"><Annotation Term="V.T"/></Singleton><ActionImport Name="Create"><Annotation Term="V.T"/></ActionImport></EntityContainer></Schema>`,
		},
		{
			name:      "expression annotations do not preserve model annotations",
			document:  `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer"><Record><Annotation Term="V.Record"/></Record></Annotation><Annotation Term="V.Record"/></EntityType></Schema>`,
			baselines: []string{`<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer"><Record><Annotation Term="V.Record"/></Record></Annotation></EntityType></Schema>`},
			expected:  `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer"><Record/></Annotation></EntityType></Schema>`,
		},
		{
			name:      "unbound action ignores parameter types in target",
			document:  `<Schema Namespace="Demo"><Action Name="Run"><Parameter Name="input" Type="Demo.Product"><Annotation Term="V.Param"/></Parameter><Annotation Term="V.T"/></Action></Schema>`,
			baselines: []string{`<Schema><Annotations Target="Demo.Run()"><Annotation Term="V.T"/></Annotations><Annotations Target="Demo.Run()/input"><Annotation Term="V.Param"/></Annotations></Schema>`},
			expected:  `<Schema Namespace="Demo"><Action Name="Run"><Parameter Name="input" Type="Demo.Product"><Annotation Term="V.Param"/></Parameter><Annotation Term="V.T"/></Action></Schema>`,
		},
		{
			name:     "global prefix preserves every qualifier",
			document: `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T"/><Annotation Term="V.T" Qualifier="A"/><Annotation Term="V.T" Qualifier="B"/><Annotation Term="Other.T"/></EntityType></Schema>`,
			rules:    []edmx.EDMXRule{{Kind: "glob", Value: "V.*"}},
			expected: `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T"/><Annotation Qualifier="A" Term="V.T"/><Annotation Qualifier="B" Term="V.T"/></EntityType></Schema>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CreateProcessor(tc.rules).Process(tc.document, tc.baselines...)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertXMLEquals(t, tc.expected, got)
		})
	}
}
