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

func TestIntegrationCSNProcessorBaselineRules(t *testing.T) {
	rules := loadRuleset(t, "testdata/integration/no_rules.json")
	input := testutils.LoadFixture("testdata/airline.csn.json")
	baseline := testutils.LoadFixture("testdata/integration/airline_baseline.json")

	document := oj.MustParseString(CreateProcessor(rules.CSN.Options, rules.CSN.Rules).Process(input, baseline)).(map[string]any)
	service := testutils.Get(t, document, "definitions", "AirlineService").(map[string]any)
	airline := testutils.Get(t, document, "definitions", "AirlineService.Airline").(map[string]any)
	airlineID := testutils.Get(t, airline, "elements", "AirlineID").(map[string]any)
	airportID := testutils.Get(t, document, "definitions", "AirlineService.Airport", "elements", "AirportID").(map[string]any)

	if got := service["__private"]; got != "retain when explicitly allowed" {
		t.Errorf("service __private = %v, want baseline-selected value", got)
	}
	if _, exists := airline["@ObjectModel.modelingPattern"]; !exists {
		t.Error("baseline-selected entity annotation was removed")
	}
	if _, exists := airline["@EndUserText.label"]; exists {
		t.Error("entity annotation absent from baseline was retained")
	}
	if _, exists := airlineID["@EndUserText.label"]; !exists {
		t.Error("baseline-selected element annotation was removed")
	}
	if _, exists := airlineID["__private"]; !exists {
		t.Error("baseline-selected element private property was removed")
	}
	if _, exists := airportID["@EndUserText.label"]; exists {
		t.Error("element annotation leaked to the same key at another path")
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
