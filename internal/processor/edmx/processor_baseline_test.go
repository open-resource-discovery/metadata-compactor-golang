//go:build unit

package edmx

import (
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

// Shared XML covers inline/external target equivalence and exact qualifier matching.
// Expected XML is retained from Node PR #5, rather than built from processor output.
func TestEDMXProcessorBaselineQualifiers(t *testing.T) {
	inputs := []struct {
		name, content string
	}{
		{"inline", `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T" String="unqualified"/><Annotation Term="V.T" Qualifier="A" String="A"/><Annotation Term="V.T" Qualifier="B" String="B"/></EntityType></Schema>`},
		{"external", `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="V.T" String="unqualified"/><Annotation Term="V.T" Qualifier="A" String="A"/><Annotation Term="V.T" Qualifier="B" String="B"/></Annotations></Schema>`},
	}
	qualifiers := []struct {
		name                             string
		inlineBaseline, externalBaseline string
		expectedInline, expectedExternal string
	}{
		{
			name:             "unqualified",
			inlineBaseline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T"/></EntityType></Schema>`,
			externalBaseline: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="V.T"/></Annotations></Schema>`,
			expectedInline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation String="unqualified" Term="V.T"/></EntityType></Schema>`,
			expectedExternal: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation String="unqualified" Term="V.T"/></Annotations></Schema>`,
		},
		{
			name:             "A",
			inlineBaseline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T" Qualifier="A"/></EntityType></Schema>`,
			externalBaseline: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="V.T" Qualifier="A"/></Annotations></Schema>`,
			expectedInline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Qualifier="A" String="A" Term="V.T"/></EntityType></Schema>`,
			expectedExternal: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Qualifier="A" String="A" Term="V.T"/></Annotations></Schema>`,
		},
		{
			name:             "B",
			inlineBaseline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="V.T" Qualifier="B"/></EntityType></Schema>`,
			externalBaseline: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Term="V.T" Qualifier="B"/></Annotations></Schema>`,
			expectedInline:   `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Qualifier="B" String="B" Term="V.T"/></EntityType></Schema>`,
			expectedExternal: `<Schema Namespace="Demo"><Annotations Target="Demo.Product"><Annotation Qualifier="B" String="B" Term="V.T"/></Annotations></Schema>`,
		},
	}
	for _, input := range inputs {
		for _, qualifier := range qualifiers {
			baselines := []struct {
				name, content string
			}{
				{"inline", qualifier.inlineBaseline},
				{"external", qualifier.externalBaseline},
			}
			for _, baseline := range baselines {
				t.Run(input.name+"/"+baseline.name+"/"+qualifier.name, func(t *testing.T) {
					expected := qualifier.expectedInline
					if input.name == "external" {
						expected = qualifier.expectedExternal
					}
					got, err := CreateProcessor(nil).Process(input.content, baseline.content)
					if err != nil {
						t.Fatal(err)
					}
					testutils.AssertXMLEquals(t, expected, got)
				})
			}
		}
	}
}
