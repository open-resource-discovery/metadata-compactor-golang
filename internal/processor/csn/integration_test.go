//go:build integration

package csn

import (
	"encoding/json"
	"testing"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

func TestIntegrationCSNProcessorAirline(t *testing.T) {
	tests := []struct {
		name         string
		rulesFixture string
		expectedFile string
	}{
		{
			name:         "example annotation allowlist",
			rulesFixture: "testdata/integration/rules.json",
			expectedFile: "testdata/integration/airline_example_rules_expected.json",
		},
		{
			name:         "empty annotation allowlist",
			rulesFixture: "testdata/integration/no_rules.json",
			expectedFile: "testdata/integration/airline_no_rules_expected.json",
		},
		{
			name:         "private property allowlist",
			rulesFixture: "testdata/integration/private_property_allowlist.json",
			expectedFile: "testdata/integration/airline_private_property_expected.json",
		},
	}

	input := testutils.LoadFixture("testdata/airline.csn.json")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRuleset(t, tc.rulesFixture)
			got := airlineDefinitions(oj.MustParseString(CreateProcessor(rules.CSN.Options, rules.CSN.Rules).Process(input)).(map[string]any))
			want := testutils.UnmarshalFixture[map[string]any](tc.expectedFile)

			testutils.AssertDeepEquals(t, want, got)
		})
	}
}

func TestIntegrationCSNProcessorAssociationOptions(t *testing.T) {
	tests := []struct {
		name         string
		rulesFixture string
		expectedFile string
	}{
		{"removes associations by default", "testdata/integration/no_rules.json", "testdata/integration/airline_associations_removed_expected.json"},
		{"preserves associations when enabled", "testdata/integration/private_property_allowlist.json", "testdata/integration/airline_associations_preserved_expected.json"},
	}

	input := testutils.LoadFixture("testdata/airline.csn.json")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRuleset(t, tc.rulesFixture)
			document := oj.MustParseString(CreateProcessor(rules.CSN.Options, rules.CSN.Rules).Process(input)).(map[string]any)

			testutils.AssertDeepEquals(t, testutils.UnmarshalFixture[map[string]any](tc.expectedFile), airportAssociations(t, document))
		})
	}
}

func TestIntegrationCSNProcessorPreserveTypesOption(t *testing.T) {
	tests := []struct {
		name         string
		rulesFixture string
		expectedFile string
	}{
		{"resolves and removes types by default", "testdata/integration/no_rules.json", "testdata/integration/airline_types_removed_expected.json"},
		{"retains custom types when enabled", "testdata/integration/preserve_types.json", "testdata/integration/airline_types_preserved_expected.json"},
	}

	input := testutils.LoadFixture("testdata/airline.csn.json")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRuleset(t, tc.rulesFixture)
			document := oj.MustParseString(CreateProcessor(rules.CSN.Options, rules.CSN.Rules).Process(input)).(map[string]any)

			testutils.AssertDeepEquals(t, testutils.UnmarshalFixture[map[string]any](tc.expectedFile), airlineTypeDefinition(t, document))
		})
	}
}

func airlineTypeDefinition(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	definitions := testutils.Get(t, document, "definitions").(map[string]any)
	result := map[string]any{
		"AirlineID": testutils.Get(t, document, "definitions", "AirlineService.Airline", "elements", "AirlineID"),
	}
	if typeDefinition, exists := definitions["AirlineUuid"]; exists {
		result["AirlineUuid"] = typeDefinition
	}
	return result
}

func airportAssociations(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	elements := testutils.Get(t, document, "definitions", "AirlineService.Airport", "elements").(map[string]any)
	result := map[string]any{}
	if association, exists := elements["to_CountryCode"]; exists {
		result["to_CountryCode"] = association
	}
	return result
}

func airlineDefinitions(document map[string]any) map[string]any {
	definitions := document["definitions"].(map[string]any)
	return map[string]any{
		"AirlineService":         definitions["AirlineService"],
		"AirlineService.Airline": definitions["AirlineService.Airline"],
		"UnassignedEntity":       definitions["UnassignedEntity"],
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
