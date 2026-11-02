//go:build unit

package metadata_compactor_golang

import (
	"fmt"
	"testing"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/csn"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

func TestProcessorProcessCSN(t *testing.T) {
	processor := Create(&model.Ruleset{
		CSN: csn.CSNRuleset{
			Options: csn.CSNOptions{PreserveTypes: true, PreserveAssociations: true},
			Rules:   []csn.CSNRule{{Kind: "exact", Value: "@Keep"}},
		},
	})

	result := oj.MustParseString(processor.Process(CSN, `{
		"definitions": {
			"CustomType": {"kind": "type", "type": "cds.String"},
			"Entity": {
				"kind": "entity",
				"@Keep": true,
				"@Remove": true,
				"elements": {
					"derived": {"type": "CustomType"},
					"association": {"type": "cds.Association", "target": "Entity"}
				}
			}
		}
	}`)).(map[string]any)

	definitions := result["definitions"].(map[string]any)
	if _, exists := definitions["CustomType"]; !exists {
		t.Error("custom type definition was not preserved")
	}

	entity := definitions["Entity"].(map[string]any)
	if _, exists := entity["@Keep"]; !exists {
		t.Error("allowed annotation was removed")
	}
	if _, exists := entity["@Remove"]; exists {
		t.Error("disallowed annotation was retained")
	}

	elements := entity["elements"].(map[string]any)
	if got := elements["derived"].(map[string]any)["type"]; got != "CustomType" {
		t.Errorf("derived type = %v, want CustomType", got)
	}
	if _, exists := elements["association"]; !exists {
		t.Error("association was not preserved")
	}
}

func TestProcessorProcessCSNWithBaseline(t *testing.T) {
	processor := Create(&model.Ruleset{})
	input := `{"definitions":{"Entity":{"kind":"entity","@Keep":true,"@Remove":true,"elements":{}}}}`
	baseline := `{"definitions":{"Entity":{"kind":"entity","@Keep":true}}}`

	result := oj.MustParseString(processor.Process(CSN, input, baseline)).(map[string]any)
	entity := result["definitions"].(map[string]any)["Entity"].(map[string]any)
	if _, exists := entity["@Keep"]; !exists {
		t.Error("baseline annotation was removed")
	}
	if _, exists := entity["@Remove"]; exists {
		t.Error("annotation absent from baseline was retained")
	}
}

func TestProcessorProcessOASWithBaseline(t *testing.T) {
	processor := Create(&model.Ruleset{})
	input := `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0","x-keep":"current","x-remove":true},"paths":{}}`
	baseline := `{"info":{"x-keep":"old"}}`
	expected := `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0","x-keep":"current"},"paths":{}}`
	result, err := processor.ProcessE(OAS, input, baseline)
	if err != nil {
		t.Fatal(err)
	}
	testutils.AssertDeepEquals(t, oj.MustParseString(expected), oj.MustParseString(result))
}

func TestProcessorProcessEDMX(t *testing.T) {
	processor := Create(&model.Ruleset{
		EDMX: edmx.EDMXRuleset{Rules: []edmx.EDMXRule{{Kind: "exact", Value: "Global"}}},
	})
	input := `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="Global"/><Annotation Term="Baseline" String="current"/><Annotation Term="Remove"/></EntityType></Schema>`
	baseline := `<Schema><Annotations Target="Demo.Product"><Annotation Term="Baseline" String="old"/></Annotations></Schema>`
	expected := `<Schema Namespace="Demo"><EntityType Name="Product"><Annotation Term="Global"/><Annotation Term="Baseline" String="current"/></EntityType></Schema>`
	testutils.AssertXMLEquals(t, expected, processor.Process(EDMX, input, baseline))
}

func TestProcessorProcessEEDMXReturnsErrors(t *testing.T) {
	output, err := Create(&model.Ruleset{}).ProcessE(EDMX, "<unclosed>")
	if err == nil || output != "" {
		t.Fatalf("expected error without partial output, got %q, %v", output, err)
	}
}

func TestProcessorProcessEDMXPanicsOnError(t *testing.T) {
	defer testutils.AssertPanics(t, "Process should panic on invalid XML")
	Create(&model.Ruleset{}).Process(EDMX, "<unclosed>")
}

func TestProcessorProcessPanicsForUnsupportedFormat(t *testing.T) {
	processor := Create(&model.Ruleset{})

	defer func() {
		if got := fmt.Sprint(recover()); got != "Unsupported format: 99" {
			t.Errorf("panic = %q, want %q", got, "Unsupported format: 99")
		}
	}()

	processor.Process(Format(99), `{}`)
}
