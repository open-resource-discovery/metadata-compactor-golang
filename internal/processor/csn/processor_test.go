//go:build unit

package csn

import (
	"testing"

	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/testutils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

func TestCSNProcessorProcess(t *testing.T) {
	t.Run("prunes disallowed properties and resolves custom types", func(t *testing.T) {
		processor := CreateProcessor(model.CSNOptions{}, []model.CSNRule{{Kind: "exact", Value: "@Keep"}})

		result := processor.Process(`{
			"definitions": {
				"CustomType": {
					"kind": "type",
					"type": "cds.String",
					"length": 42
				},
				"Context": {
					"kind": "context",
					"@Keep": true,
					"@Remove": true,
					"__private": true,
					"name": "Context"
				},
				"Service": {
					"kind": "service",
					"@Remove": true,
					"__private": true,
					"name": "Service"
				},
				"Entity": {
					"kind": "entity",
					"@Remove": true,
					"__private": true,
					"elements": {
						"derived": {
							"type": "CustomType",
							"@Keep": true,
							"@Remove": true,
							"__private": true
						},
						"overridden": {
							"type": "CustomType",
							"length": 10
						},
						"builtin": {
							"type": "cds.Integer"
						}
					}
				}
			}
		}`)

		document := oj.MustParseString(result).(map[string]any)
		definitions := testutils.Get(t, document, "definitions").(map[string]any)
		if _, exists := definitions["CustomType"]; exists {
			t.Error("custom type definition was not removed")
		}

		assertKeys(t, testutils.Get(t, document, "definitions", "Context").(map[string]any), "kind", "@Keep", "name")
		assertKeys(t, testutils.Get(t, document, "definitions", "Service").(map[string]any), "kind", "name")

		entity := testutils.Get(t, document, "definitions", "Entity").(map[string]any)
		assertKeys(t, entity, "kind", "elements")
		derived := testutils.Get(t, document, "definitions", "Entity", "elements", "derived").(map[string]any)
		assertKeys(t, derived, "type", "length", "@Keep")
		if got := derived["type"]; got != "cds.String" {
			t.Errorf("derived type = %v, want cds.String", got)
		}
		if got := derived["length"]; got != int64(42) {
			t.Errorf("derived length = %v, want 42", got)
		}
		if _, exists := derived["kind"]; exists {
			t.Error("custom type metadata was copied to the element")
		}

		overridden := testutils.Get(t, document, "definitions", "Entity", "elements", "overridden").(map[string]any)
		if got := overridden["length"]; got != int64(10) {
			t.Errorf("overridden length = %v, want 10", got)
		}
		if got := overridden["type"]; got != "cds.String" {
			t.Errorf("overridden type = %v, want cds.String", got)
		}

		builtin := testutils.Get(t, document, "definitions", "Entity", "elements", "builtin").(map[string]any)
		if got := builtin["type"]; got != "cds.Integer" {
			t.Errorf("builtin type = %v, want cds.Integer", got)
		}
	})

	t.Run("removes all annotations and private properties when no rules are supplied", func(t *testing.T) {
		result := CreateProcessor(model.CSNOptions{}, nil).Process(`{
			"definitions": {
				"Entity": {
					"kind": "entity",
					"@Annotation": true,
					"__private": true,
					"elements": {
						"field": {"type": "cds.String", "@Annotation": true, "__private": true}
					}
				}
			}
		}`)

		document := oj.MustParseString(result).(map[string]any)
		entity := testutils.Get(t, document, "definitions", "Entity").(map[string]any)
		assertKeys(t, entity, "kind", "elements")
		field := testutils.Get(t, document, "definitions", "Entity", "elements", "field").(map[string]any)
		assertKeys(t, field, "type")
	})
}

func TestCSNProcessorBranchCases(t *testing.T) {
	t.Run("glob rule keeps matching annotations only", func(t *testing.T) {
		processor := CreateProcessor(model.CSNOptions{}, []model.CSNRule{{Kind: "glob", Value: "@PersonalData.*"}})

		result := processor.Process(`{
			"definitions": {
				"Entity": {
					"kind": "entity",
					"@PersonalData.entitySemantics": true,
					"@EndUserText.label": "remove",
					"elements": {}
				}
			}
		}`)

		entity := oj.MustParseString(result).(map[string]any)["definitions"].(map[string]any)["Entity"].(map[string]any)
		assertKeys(t, entity, "kind", "@PersonalData.entitySemantics", "elements")
	})

	t.Run("leaves builtin, untyped, and undefined custom types unresolved", func(t *testing.T) {
		result := CreateProcessor(model.CSNOptions{PreserveTypes: true}, nil).Process(`{
			"definitions": {
				"Entity": {
					"kind": "entity",
					"elements": {
						"builtin": {"type": "cds.String"},
						"undefined": {"type": "MissingType"},
						"nonStringType": {"type": 1},
						"untyped": {}
					}
				}
			}
		}`)

		elements := oj.MustParseString(result).(map[string]any)["definitions"].(map[string]any)["Entity"].(map[string]any)["elements"].(map[string]any)
		if got := elements["builtin"].(map[string]any)["type"]; got != "cds.String" {
			t.Errorf("builtin type = %v, want cds.String", got)
		}
		if got := elements["undefined"].(map[string]any)["type"]; got != "MissingType" {
			t.Errorf("undefined type = %v, want MissingType", got)
		}
		if got := elements["nonStringType"].(map[string]any)["type"]; got != int64(1) {
			t.Errorf("non-string type = %v, want 1", got)
		}
		assertKeys(t, elements["untyped"].(map[string]any))
	})
}

func TestCSNProcessorOptions(t *testing.T) {
	tests := []struct {
		name                  string
		options               model.CSNOptions
		wantTypeDefinition    bool
		wantResolvedType      string
		wantAssociationExists bool
	}{
		{"removes types and associations by default", model.CSNOptions{}, false, "cds.String", false},
		{"preserves type definitions", model.CSNOptions{PreserveTypes: true}, true, "CustomType", false},
		{"preserves associations", model.CSNOptions{PreserveAssociations: true}, false, "cds.String", true},
		{"preserves types and associations", model.CSNOptions{PreserveTypes: true, PreserveAssociations: true}, true, "CustomType", true},
	}

	const document = `{
		"definitions": {
			"CustomType": {"kind": "type", "type": "cds.String", "length": 20, "@Keep": true, "@Remove": true, "__private": true},
			"Entity": {
				"kind": "entity",
				"elements": {
					"derived": {"type": "CustomType"},
					"association": {"type": "cds.Association", "target": "Entity"}
				}
			}
		}
	}`

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			definitions := oj.MustParseString(CreateProcessor(tc.options, []model.CSNRule{{Kind: "exact", Value: "@Keep"}}).Process(document)).(map[string]any)["definitions"].(map[string]any)
			typeDefinition, hasTypeDefinition := definitions["CustomType"]
			if hasTypeDefinition != tc.wantTypeDefinition {
				t.Errorf("custom type definition present = %v, want %v", hasTypeDefinition, tc.wantTypeDefinition)
			}
			if hasTypeDefinition {
				assertKeys(t, typeDefinition.(map[string]any), "kind", "type", "length", "@Keep")
			}

			elements := definitions["Entity"].(map[string]any)["elements"].(map[string]any)
			if got := elements["derived"].(map[string]any)["type"]; got != tc.wantResolvedType {
				t.Errorf("derived type = %v, want %s", got, tc.wantResolvedType)
			}
			_, hasAssociation := elements["association"]
			if hasAssociation != tc.wantAssociationExists {
				t.Errorf("association present = %v, want %v", hasAssociation, tc.wantAssociationExists)
			}
		})
	}
}

func assertKeys(t *testing.T, value map[string]any, want ...string) {
	t.Helper()
	if len(value) != len(want) {
		t.Fatalf("got keys %v, want %v", mapKeys(value), want)
	}
	for _, key := range want {
		if _, exists := value[key]; !exists {
			t.Errorf("missing key %q; got keys %v", key, mapKeys(value))
		}
	}
}

func mapKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	return keys
}

func assertPanics(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	call()
}
