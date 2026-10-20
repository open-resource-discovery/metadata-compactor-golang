//go:build unit

package oas

import (
	"encoding/json"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/jputils"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
)

func TestOASRuleMatches(t *testing.T) {
	tests := []struct {
		name  string
		rule  OASRule
		value string
		want  bool
	}{
		{"exact rule matches the same value", OASRule{Kind: "exact", Value: "x-internal"}, "x-internal", true},
		{"exact rule does not match a different value", OASRule{Kind: "exact", Value: "x-internal"}, "x-public", false},
		{"glob rule matches its base value", OASRule{Kind: "glob", Value: "x-internal-*"}, "x-internal", true},
		{"glob rule matches a suffixed value", OASRule{Kind: "glob", Value: "x-internal-*"}, "x-internal-id", true},
		{"glob rule does not match a value without the hyphen boundary", OASRule{Kind: "glob", Value: "x-internal-*"}, "x-internalized", false},
		{"glob rule does not match another namespace", OASRule{Kind: "glob", Value: "x-internal-*"}, "x-public-id", false},
		{"unrecognized rule kind does not match", OASRule{Kind: "unknown", Value: "x-internal"}, "x-internal", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rule.Matches(jputils.Expr("$"), tc.value); got != tc.want {
				t.Errorf("Matches(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestOASRuleMatchesPath(t *testing.T) {
	rule := OASRule{Kind: "exact", Value: "x-keep", Path: "$.components.schemas.Entity['x-keep']"}

	if !rule.Matches(jputils.Expr("$", "components", "schemas", "Entity", "x-keep"), "x-keep") {
		t.Error("path-specific rule did not match its configured path")
	}
	if rule.Matches(jputils.Expr("$", "components", "schemas", "Other", "x-keep"), "x-keep") {
		t.Error("path-specific rule matched a different path")
	}
}

func TestOASRuleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		want      OASRule
		wantError bool
	}{
		{"exact rule", `"x-internal"`, OASRule{Kind: "exact", Value: "x-internal"}, false},
		{"glob rule", `"x-internal-*"`, OASRule{Kind: "glob", Value: "x-internal-*"}, false},
		{"asterisk without hyphen is exact", `"x-internal*"`, OASRule{Kind: "exact", Value: "x-internal*"}, false},
		{"invalid JSON", `not-json`, OASRule{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var rule OASRule
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
