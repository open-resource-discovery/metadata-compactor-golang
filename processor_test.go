//go:build unit

package metadata_compactor_golang

import (
	"fmt"
	"testing"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

func TestProcessorProcessCSN(t *testing.T) {
	processor := Create(&model.Ruleset{
		CSN: model.CSNRuleset{
			Options: model.CSNOptions{PreserveTypes: true, PreserveAssociations: true},
			Rules:   []model.CSNRule{{Kind: "exact", Value: "@Keep"}},
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

func TestProcessorProcessPanicsForUnsupportedFormat(t *testing.T) {
	processor := Create(&model.Ruleset{})

	defer func() {
		if got := fmt.Sprint(recover()); got != "Unsupported format: 99" {
			t.Errorf("panic = %q, want %q", got, "Unsupported format: 99")
		}
	}()

	processor.Process(Format(99), `{}`)
}
