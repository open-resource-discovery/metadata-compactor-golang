//go:build integration

package edmx

import (
	"encoding/json"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

// Expected XML comes from Node PR #5 at 9eada5f8fdb340b70d7173179e8b87fcbd53175c.
func TestIntegrationEDMXProcessorEnvelope(t *testing.T) {
	tests := []struct {
		name, rulesFixture, expectedFile string
	}{
		{"exact annotation allowlist", "testdata/integration/global_exact_rules.json", "testdata/integration/full_edmx_envelope_expected.edmx"},
		{"prefix annotation allowlist", "testdata/integration/prefix_rules.json", "testdata/integration/full_edmx_prefix_rules_expected.edmx"},
		{"empty annotation allowlist", "testdata/integration/no_rules.json", "testdata/integration/full_edmx_no_rules_expected.edmx"},
	}
	input := testutils.LoadFixture("testdata/full_edmx_envelope.edmx")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRuleset(t, tc.rulesFixture)
			got, err := CreateProcessor(rules.EDMX.Rules).Process(input)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertXMLEquals(t, testutils.LoadFixture(tc.expectedFile), got)
		})
	}
}

func TestIntegrationEDMXProcessorBaselines(t *testing.T) {
	tests := []struct {
		name             string
		inputFixture     string
		rulesFixture     string
		expectedFile     string
		baselineFixtures []string
	}{
		{
			name:             "external baseline matches inline property",
			inputFixture:     "testdata/external_baseline_matches_inline_property.edmx",
			rulesFixture:     "testdata/integration/no_rules.json",
			expectedFile:     "testdata/integration/external_baseline_matches_inline_property_expected.edmx",
			baselineFixtures: []string{"testdata/integration/external_baseline_matches_inline_property_baseline_1.edmx"},
		},
		{
			name:             "inline baseline matches external property",
			inputFixture:     "testdata/inline_baseline_matches_external_property.edmx",
			rulesFixture:     "testdata/integration/no_rules.json",
			expectedFile:     "testdata/integration/inline_baseline_matches_external_property_expected.edmx",
			baselineFixtures: []string{"testdata/integration/inline_baseline_matches_external_property_baseline_1.edmx"},
		},
		{
			name:             "multiple baselines combine with globals",
			inputFixture:     "testdata/multiple_baselines_combine_with_globals.edmx",
			rulesFixture:     "testdata/integration/multiple_baselines_combine_with_globals_rules.json",
			expectedFile:     "testdata/integration/multiple_baselines_combine_with_globals_expected.edmx",
			baselineFixtures: []string{"testdata/integration/multiple_baselines_combine_with_globals_baseline_1.edmx", "testdata/integration/multiple_baselines_combine_with_globals_baseline_2.edmx"},
		},
		{
			name:             "bound action binding types",
			inputFixture:     "testdata/operation_overload_signatures.edmx",
			rulesFixture:     "testdata/integration/no_rules.json",
			expectedFile:     "testdata/integration/operation_overload_signatures_expected.edmx",
			baselineFixtures: []string{"testdata/integration/operation_overload_signatures_baseline_1.edmx"},
		},
		{
			name:             "bound action nested annotation targets",
			inputFixture:     "testdata/bound_action_nested_annotation_targets.edmx",
			rulesFixture:     "testdata/integration/no_rules.json",
			expectedFile:     "testdata/integration/bound_action_nested_annotation_targets_expected.edmx",
			baselineFixtures: []string{"testdata/integration/bound_action_nested_annotation_targets_baseline_1.edmx"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRuleset(t, tc.rulesFixture)
			baselines := make([]string, 0, len(tc.baselineFixtures))
			for _, path := range tc.baselineFixtures {
				baselines = append(baselines, testutils.LoadFixture(path))
			}
			got, err := CreateProcessor(rules.EDMX.Rules).Process(testutils.LoadFixture(tc.inputFixture), baselines...)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertXMLEquals(t, testutils.LoadFixture(tc.expectedFile), got)
		})
	}
}

func loadRuleset(t *testing.T, path string) model.Ruleset {
	t.Helper()
	var rules model.Ruleset
	if err := json.Unmarshal([]byte(testutils.LoadFixture(path)), &rules); err != nil {
		t.Fatalf("load ruleset %q: %v", path, err)
	}
	return rules
}
