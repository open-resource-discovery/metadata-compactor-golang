//go:build unit

package edmx

import (
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

func TestEDMXRuleExtractorTargets(t *testing.T) {
	baseline := `<Schema Namespace="Demo"><EntityType Name="Product"><Property Name="ID"><Annotation Term="Property.Term"/></Property></EntityType><Action Name="Create"><Parameter Name="input" Type="Demo.Product"><Annotation Term="Parameter.Term"/></Parameter><ReturnType Type="Demo.Product"><Annotation Term="Return.Term"/></ReturnType><Annotation Term="Action.Term"/></Action><Annotations Target="Demo.External"><Annotation Term="External.Term" Qualifier="Summary"/></Annotations></Schema>`
	got, err := (RuleExtractor{}).Extract(baseline)
	if err != nil {
		t.Fatal(err)
	}
	targets := []string{"Demo.Product/ID", "Demo.Create()/input", "Demo.Create()/$ReturnType", "Demo.Create()", "Demo.External"}
	unqualified, summary := "", "Summary"
	want := []edmx.EDMXRule{
		{Kind: "exact", Value: "Property.Term", Target: &targets[0], Qualifier: &unqualified},
		{Kind: "exact", Value: "Parameter.Term", Target: &targets[1], Qualifier: &unqualified},
		{Kind: "exact", Value: "Return.Term", Target: &targets[2], Qualifier: &unqualified},
		{Kind: "exact", Value: "Action.Term", Target: &targets[3], Qualifier: &unqualified},
		{Kind: "exact", Value: "External.Term", Target: &targets[4], Qualifier: &summary},
	}
	testutils.AssertDeepEquals(t, want, got)
}
