//go:build unit

package edmx

import "testing"

func TestEDMXRuleMatches(t *testing.T) {
	target := "Demo.Product/ID"
	empty := ""
	summary := "Summary"
	tests := []struct {
		name                    string
		rule                    EDMXRule
		target, term, qualifier string
		want                    bool
	}{
		{"exact", EDMXRule{Kind: "exact", Value: "Core.Description"}, "Demo.Product", "Core.Description", "", true},
		{"literal at sign", EDMXRule{Kind: "exact", Value: "@Core.Description"}, "", "Core.Description", "", false},
		{"prefix", EDMXRule{Kind: "glob", Value: "Common.*"}, "", "Common.Label", "", true},
		{"dot boundary", EDMXRule{Kind: "glob", Value: "Common.*"}, "", "CommonExtra.Label", "", false},
		{"case sensitive", EDMXRule{Kind: "exact", Value: "Core.Description"}, "", "core.Description", "", false},
		{"same target", EDMXRule{Kind: "exact", Value: "One", Target: &target}, target, "One", "", true},
		{"different target", EDMXRule{Kind: "exact", Value: "One", Target: &target}, "Demo.Product/Other", "One", "", false},
		{"global rule keeps qualified annotation", EDMXRule{Kind: "exact", Value: "One"}, target, "One", summary, true},
		{"empty target is not global", EDMXRule{Kind: "exact", Value: "One", Target: &empty}, target, "One", "", false},
		{"empty target matches empty target", EDMXRule{Kind: "exact", Value: "One", Target: &empty}, "", "One", "", true},
		{"unqualified baseline matches unqualified", EDMXRule{Kind: "exact", Value: "One", Qualifier: &empty}, target, "One", "", true},
		{"unqualified baseline rejects qualified", EDMXRule{Kind: "exact", Value: "One", Qualifier: &empty}, target, "One", summary, false},
		{"qualified baseline matches same qualifier", EDMXRule{Kind: "exact", Value: "One", Qualifier: &summary}, target, "One", summary, true},
		{"qualified baseline rejects other qualifier", EDMXRule{Kind: "exact", Value: "One", Qualifier: &summary}, target, "One", "Other", false},
		{"qualified baseline rejects unqualified", EDMXRule{Kind: "exact", Value: "One", Qualifier: &summary}, target, "One", "", false},
		{"unknown kind", EDMXRule{Kind: "unknown", Value: "One"}, "", "One", "", false},
		{"invalid manual glob", EDMXRule{Kind: "glob"}, "", "One", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rule.Matches(tc.target, tc.term, tc.qualifier); got != tc.want {
				t.Fatalf("Matches = %v, want %v", got, tc.want)
			}
		})
	}
}
