//go:build integration

package oas

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

func TestIntegrationOASProcessorPetstore(t *testing.T) {
	tests := []struct {
		name            string
		rulesFixture    string
		expectedFixture string
	}{
		{
			name:            "extension family allowlist",
			rulesFixture:    "testdata/integration/rules.json",
			expectedFixture: "testdata/integration/petstore_sap_rules_expected.json",
		},
		{
			name:            "exact extension allowlist",
			rulesFixture:    "testdata/integration/exact_rule.json",
			expectedFixture: "testdata/integration/petstore_exact_rule_expected.json",
		},
		{
			name:            "empty extension allowlist",
			rulesFixture:    "testdata/integration/no_rules.json",
			expectedFixture: "testdata/integration/petstore_no_rules_expected.json",
		},
	}

	input := testutils.LoadFixture("testdata/petstore.oas.json")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadOASRuleset(t, tc.rulesFixture)
			document := oj.MustParseString(CreateProcessor(rules.OAS.Rules).Process(input))

			want := testutils.UnmarshalFixture[map[string]any](tc.expectedFixture)
			testutils.AssertDeepEquals(t, want, specificationExtensions(document))
		})
	}
}

func TestIntegrationOASProcessorPreservesStandardPetstoreContent(t *testing.T) {
	input := testutils.LoadFixture("testdata/petstore.oas.json")
	rules := loadOASRuleset(t, "testdata/integration/no_rules.json")
	got := oj.MustParseString(CreateProcessor(rules.OAS.Rules).Process(input))
	want := oj.MustParseString(input)
	removeSpecificationExtensions(want)

	testutils.AssertDeepEquals(t, want, got)
}

func TestIntegrationOASProcessorBaselineRules(t *testing.T) {
	input := testutils.LoadFixture("testdata/petstore.oas.json")
	baseline := testutils.LoadFixture("testdata/integration/petstore_baseline.json")
	rules := loadOASRuleset(t, "testdata/integration/no_rules.json")
	document := oj.MustParseString(CreateProcessor(rules.OAS.Rules).Process(input, baseline))

	want := testutils.UnmarshalFixture[map[string]any]("testdata/integration/petstore_baseline_expected.json")
	testutils.AssertDeepEquals(t, want, specificationExtensions(document))
}

func loadOASRuleset(t *testing.T, path string) model.Ruleset {
	t.Helper()

	var rules model.Ruleset
	if err := json.Unmarshal([]byte(testutils.LoadFixture(path)), &rules); err != nil {
		t.Fatalf("load ruleset %q: %v", path, err)
	}

	return rules
}

func specificationExtensions(document any) map[string]any {
	result := make(map[string]any)
	collectSpecificationExtensions(document, jp.MustParseString("$"), result)
	return result
}

func collectSpecificationExtensions(value any, path jp.Expr, result map[string]any) {
	switch value := value.(type) {
	case map[string]any:
		for name, child := range value {
			childPath := path.Child(name)
			if strings.HasPrefix(name, "x-") {
				result[childPath.String()] = child
			}
			collectSpecificationExtensions(child, childPath, result)
		}
	case []any:
		for index, child := range value {
			collectSpecificationExtensions(child, path.Nth(index), result)
		}
	}
}

func removeSpecificationExtensions(value any) {
	switch value := value.(type) {
	case map[string]any:
		for name, child := range value {
			if strings.HasPrefix(name, "x-") {
				delete(value, name)
				continue
			}
			removeSpecificationExtensions(child)
		}
	case []any:
		for _, child := range value {
			removeSpecificationExtensions(child)
		}
	}
}
