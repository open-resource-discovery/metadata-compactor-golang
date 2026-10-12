//go:build unit

package model

import (
	"encoding/json"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/jputils"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

func TestCSNRuleMatches(t *testing.T) {
	tests := []struct {
		name  string
		rule  CSNRule
		value string
		want  bool
	}{
		{"exact rule matches the same value", CSNRule{Kind: "exact", Value: "@EndUserText.label"}, "@EndUserText.label", true},
		{"exact rule does not match a different value", CSNRule{Kind: "exact", Value: "@EndUserText.label"}, "@EndUserText.heading", false},
		{"glob rule matches a child in its namespace", CSNRule{Kind: "glob", Value: "@PersonalData.*"}, "@PersonalData.entitySemantics", true},
		{"glob rule does not match another namespace", CSNRule{Kind: "glob", Value: "@PersonalData.*"}, "@EndUserText.label", false},
		{"unrecognized rule kind does not match", CSNRule{Kind: "unknown", Value: "@PersonalData.*"}, "@PersonalData.entitySemantics", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rule.Matches(jputils.Expr("$"), tc.value); got != tc.want {
				t.Errorf("Matches(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestCSNRuleMatchesPath(t *testing.T) {
	rule := CSNRule{Kind: "exact", Value: "@Keep", Path: "$.definitions.Entity['@Keep']"}

	if !rule.Matches(jputils.Expr("$", "definitions", "Entity", "@Keep"), "@Keep") {
		t.Error("path-specific rule did not match its configured path")
	}
	if rule.Matches(jputils.Expr("$", "definitions", "Other", "@Keep"), "@Keep") {
		t.Error("path-specific rule matched a different path")
	}
}

func TestCSNRuleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      CSNRule
		wantError bool
	}{
		{"exact rule", `"@EndUserText.label"`, CSNRule{Kind: "exact", Value: "@EndUserText.label"}, false},
		{"glob rule", `"@PersonalData.*"`, CSNRule{Kind: "glob", Value: "@PersonalData.*"}, false},
		{"invalid JSON", `not-json`, CSNRule{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var rule CSNRule
			err := json.Unmarshal([]byte(tc.data), &rule)
			if tc.wantError {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("json.Unmarshal() returned an error: %v", err)
			}

			testutils.AssertDeepEquals(t, tc.want, rule)
		})
	}
}
