//go:build unit

package edmx

import (
	"fmt"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/xmltree"
)

func TestAnnotationTargetResolverTargets(t *testing.T) {
	tests := []struct {
		name, content string
		expected      []string
	}{
		{"schema elements and children", `<edm:Schema xmlns:edm="urn:edm" Namespace="Demo"><edm:Annotation Term="V.T"/><edm:TypeDefinition Name="Code"><edm:Annotation Term="V.T"/></edm:TypeDefinition><edm:Term Name="Label"><edm:Annotation Term="V.T"/></edm:Term><edm:EnumType Name="Color"><edm:Annotation Term="V.T"/><edm:Member Name="Red"><edm:Annotation Term="V.T"/></edm:Member></edm:EnumType><edm:EntityType Name="Product"><edm:Annotation Term="V.T"/><edm:Property Name="ID"><edm:Annotation Term="V.T"/></edm:Property><edm:NavigationProperty Name="Category"><edm:Annotation Term="V.T"/></edm:NavigationProperty></edm:EntityType><edm:ComplexType Name="Address"><edm:Annotation Term="V.T"/><edm:Property Name="City"><edm:Annotation Term="V.T"/></edm:Property></edm:ComplexType><edm:EntityContainer Name="Service"><edm:Annotation Term="V.T"/><edm:EntitySet Name="Products"><edm:Annotation Term="V.T"/></edm:EntitySet><edm:Singleton Name="Main"><edm:Annotation Term="V.T"/></edm:Singleton><edm:ActionImport Name="Create"><edm:Annotation Term="V.T"/></edm:ActionImport><edm:FunctionImport Name="Find"><edm:Annotation Term="V.T"/></edm:FunctionImport></edm:EntityContainer></edm:Schema>`, []string{"Demo", "Demo.Code", "Demo.Label", "Demo.Color", "Demo.Color/Red", "Demo.Product", "Demo.Product/ID", "Demo.Product/Category", "Demo.Address", "Demo.Address/City", "Demo.Service", "Demo.Service/Products", "Demo.Service/Main", "Demo.Service/Create", "Demo.Service/Find"}},
		{"nested annotation targets", `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer" Qualifier="Tablet"><Annotation Term="V.Inner"><Annotation Term="V.Deep"/></Annotation></Annotation></EntityType><Annotations Target="Demo.Product/ID" Qualifier="Phone"><Annotation Term="V.Outer"><Annotation Term="V.Inner" Qualifier="Nested"><Annotation Term="V.Deep"/></Annotation></Annotation></Annotations><Annotations Target="Demo.Product/@V.Existing#Q"><Annotation Term="V.T"/></Annotations></Schema>`, []string{"Demo.Product", "Demo.Product/@V.Outer#Tablet", "Demo.Product/@V.Outer#Tablet/@V.Inner", "Demo.Product/ID", "Demo.Product/ID/@V.Outer#Phone", "Demo.Product/ID/@V.Outer#Phone/@V.Inner#Nested", "Demo.Product/@V.Existing#Q"}},
		{"expressions are not model targets", `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.Outer"><Record><Annotation Term="V.Record"/><PropertyValue Property="Label"><Annotation Term="V.Value"/></PropertyValue></Record></Annotation></EntityType></Schema>`, []string{"Demo.Product", "", ""}},
		{"zero parameter function", `<Schema Namespace="Demo"><Function Name="Ping"><Annotation Term="V.T"/></Function></Schema>`, []string{"Demo.Ping()"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { testutils.AssertDeepEquals(t, tc.expected, annotationTargets(t, tc.content)) })
	}
}

func TestAnnotationTargetResolverOperations(t *testing.T) {
	tests := []struct{ name, element, attributes, signature string }{
		{"unbound action", "Action", "", "Demo.Run()"},
		{"explicit unbound action", "Action", `IsBound="false"`, "Demo.Run()"},
		{"bound action", "Action", `IsBound="true"`, "Demo.Run(Collection(Demo.Product))"},
		{"numeric bound action", "Action", `IsBound="1"`, "Demo.Run(Collection(Demo.Product))"},
		{"unbound function", "Function", "", "Demo.Run(Collection(Demo.Product),Edm.String)"},
		{"bound function", "Function", `IsBound="true"`, "Demo.Run(Collection(Demo.Product),Edm.String)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content := fmt.Sprintf(`<Schema Namespace="Demo"><%s Name="Run" %s><Annotation Term="V.T"/><Parameter Name="binding" Type="Collection(Demo.Product)"><Annotation Term="V.T"/></Parameter><Parameter Name="query" Type="Edm.String"><Annotation Term="V.T"/></Parameter><ReturnType Type="Edm.String"><Annotation Term="V.T"/></ReturnType></%s></Schema>`, tc.element, tc.attributes, tc.element)
			expected := []string{tc.signature, tc.signature + "/binding", tc.signature + "/query", tc.signature + "/$ReturnType"}
			testutils.AssertDeepEquals(t, expected, annotationTargets(t, content))
		})
	}
}

func TestAnnotationTargetResolverUnresolvedTargets(t *testing.T) {
	resolver := AnnotationTargetResolver{}
	for _, chain := range [][]xmltree.Node{nil, {{"name": "Property"}}, {{"name": "Annotation"}}} {
		if target, found := resolver.Resolve(chain); found {
			t.Fatalf("unexpected target %q", target)
		}
	}
	tests := []struct{ name, content string }{
		{"missing namespace", `<Schema><EntityType Name="Product"><Annotation Term="V.T"/></EntityType></Schema>`},
		{"missing type name", `<Schema Namespace="Demo"><EntityType><Property Name="ID"><Annotation Term="V.T"/></Property></EntityType></Schema>`},
		{"missing property name", `<Schema Namespace="Demo"><EntityType Name="Product"><Property><Annotation Term="V.T"/></Property></EntityType></Schema>`},
		{"missing external target", `<Schema Namespace="Demo"><Annotations><Annotation Term="V.T"/></Annotations></Schema>`},
		{"empty external target", `<Schema Namespace="Demo"><Annotations Target=""><Annotation Term="V.T"/></Annotations></Schema>`},
		{"whitespace external target", `<Schema Namespace="Demo"><Annotations Target="&#xFEFF;"><Annotation Term="V.T"/></Annotations></Schema>`},
		{"missing function parameter type", `<Schema Namespace="Demo"><Function Name="Run"><Parameter Name="input"><Annotation Term="V.T"/></Parameter></Function></Schema>`},
		{"missing binding parameter", `<Schema Namespace="Demo"><Action Name="Run" IsBound="true"><Annotation Term="V.T"/></Action></Schema>`},
		{"missing binding parameter type", `<Schema Namespace="Demo"><Action Name="Run" IsBound="true"><Parameter Name="binding"/><Annotation Term="V.T"/></Action></Schema>`},
		{"bound function without parameters", `<Schema Namespace="Demo"><Function Name="Run" IsBound="true"><Annotation Term="V.T"/></Function></Schema>`},
		{"parameter outside operation", `<Schema Namespace="Demo"><Parameter Name="input"><Annotation Term="V.T"/></Parameter></Schema>`},
		{"unknown parent", `<Schema Namespace="Demo"><EntityType Name="Product"><Unknown><Annotation Term="V.T"/></Unknown></EntityType></Schema>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.AssertDeepEquals(t, []string{""}, annotationTargets(t, tc.content))
			rules, err := (RuleExtractor{}).Extract(tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if len(rules) != 0 {
				t.Fatalf("unresolved baseline produced rules: %#v", rules)
			}
		})
	}
}

func TestAnnotationTargetResolverQualifier(t *testing.T) {
	content := `<Schema Namespace="Demo"><Annotations Target="Demo.Product" Qualifier="A"><Annotation Term="V.Outer"><Annotation Term="V.Inner"/><Annotation Term="V.Inner" Qualifier="B"/></Annotation><Annotation Term="V.Empty" Qualifier=""/></Annotations><EntityType Name="Product"><Annotation Term="V.Unqualified"/></EntityType></Schema>`
	expected := []struct {
		value string
		found bool
	}{{"A", true}, {"", false}, {"B", true}, {"", true}, {"", false}}
	contexts := annotationContexts(t, content)
	if len(contexts) != len(expected) {
		t.Fatalf("got %d contexts, want %d", len(contexts), len(expected))
	}
	resolver := AnnotationTargetResolver{}
	for index, context := range contexts {
		qualifier, found := resolver.ResolveQualifier(context)
		if qualifier != expected[index].value || found != expected[index].found {
			t.Errorf("annotation %d: qualifier %q, found %v", index, qualifier, found)
		}
	}
	if _, found := resolver.ResolveQualifier(nil); found {
		t.Error("empty context has a qualifier")
	}
}

func annotationTargets(t *testing.T, content string) []string {
	t.Helper()
	var targets []string
	for _, context := range annotationContexts(t, content) {
		target, _ := (AnnotationTargetResolver{}).Resolve(context)
		targets = append(targets, target)
	}
	return targets
}

func annotationContexts(t *testing.T, content string) [][]xmltree.Node {
	t.Helper()
	document, err := xmltree.Convert(content)
	if err != nil {
		t.Fatal(err)
	}
	var result [][]xmltree.Node
	var visit func([]xmltree.Node, []xmltree.Node)
	visit = func(nodes, parents []xmltree.Node) {
		for _, node := range nodes {
			if node.Kind() != "element" {
				continue
			}
			chain := append(parents, node)
			if localName(node) == "Annotation" {
				result = append(result, append([]xmltree.Node(nil), chain...))
			}
			visit(node.Nodes(), chain)
		}
	}
	visit(document.Nodes(), nil)
	return result
}
