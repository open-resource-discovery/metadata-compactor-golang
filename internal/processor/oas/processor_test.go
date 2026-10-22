package oas

import (
	"reflect"
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	modeloas "github.com/open-resource-discovery/metadata-compactor-golang/model/oas"
)

func TestCreateProcessorCopiesRules(t *testing.T) {
	rules := []modeloas.OASRule{{Kind: "exact", Value: "x-preserved"}}
	processor := CreateProcessor(rules)
	rules[0].Value = "x-mutated"

	result := oj.MustParseString(processor.Process(`{"x-preserved":true,"x-mutated":true}`))
	if !jp.MustParseString("$['x-preserved']").Has(result) {
		t.Error("processor did not retain its copied rule")
	}
	if jp.MustParseString("$['x-mutated']").Has(result) {
		t.Error("processor rules changed when the caller mutated its input slice")
	}
}

func TestProcessAppliesExactGlobAndPathRules(t *testing.T) {
	processor := CreateProcessor([]modeloas.OASRule{
		{Kind: "exact", Value: "x-global"},
		{Kind: "glob", Value: "x-family-*"},
		{Kind: "exact", Value: "x-scoped", Path: "$.info['x-scoped']"},
	})

	result := oj.MustParseString(processor.Process(`{
		"openapi":"3.2.0",
		"regular":true,
		"x-global":1,
		"x-family-one":2,
		"x-remove":3,
		"x-scoped":4,
		"info":{"x-scoped":5}
	}`))

	for _, path := range []string{"$.regular", "$['x-global']", "$['x-family-one']", "$.info['x-scoped']"} {
		if !jp.MustParseString(path).Has(result) {
			t.Errorf("expected %s to be preserved", path)
		}
	}
	for _, path := range []string{"$['x-remove']", "$['x-scoped']"} {
		if jp.MustParseString(path).Has(result) {
			t.Errorf("expected %s to be removed", path)
		}
	}
}

func TestProcessUsesSpecificationExtensionsFromEveryBaseline(t *testing.T) {
	processor := CreateProcessor(nil)
	document := `{
		"x-root-one":1,
		"x-root-two":2,
		"x-root-remove":3,
		"info":{"title":"current","x-info":true},
		"paths":{"/pets":{"x-path-item":true}}
	}`
	baselineOne := `{
		"x-root-one":null,
		"ordinary":true,
		"info":{"x-info":null},
		"paths":{"/pets":{"x-path-item":null}}
	}`
	baselineTwo := `{"x-root-two":null}`

	got := oj.MustParseString(processor.Process(document, baselineOne, baselineTwo))
	want := oj.MustParseString(`{
		"x-root-one":1,
		"x-root-two":2,
		"info":{"title":"current","x-info":true},
		"paths":{"/pets":{"x-path-item":true}}
	}`)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected processed document:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestProcessPreservesCustomHTTPHeaderStartingWithX(t *testing.T) {
	result := oj.MustParseString(CreateProcessor(nil).Process(`{
		"paths": {
			"/pets": {
				"get": {
					"responses": {
						"200": {
							"description": "ok",
							"headers": {
								"x-request-id": {
									"description": "Request identifier",
									"x-internal": true
								}
							}
						}
					}
				}
			}
		}
	}`))

	header := jp.MustParseString(`$.paths['/pets'].get.responses['200'].headers['x-request-id']`).First(result)
	if header == nil {
		t.Fatal("custom x-request-id HTTP header was removed")
	}
	if jp.MustParseString(`$.paths['/pets'].get.responses['200'].headers['x-request-id']['x-internal']`).Has(result) {
		t.Error("specification extension inside the Header Object was not removed")
	}
}

func TestWalkSpecificationExtensionObjectsIsPrecise(t *testing.T) {
	document := map[string]any{
		"paths": map[string]any{"/pets": map[string]any{
			"get": map[string]any{
				"responses": map[string]any{"200": map[string]any{"description": "ok"}},
				"callbacks": map[string]any{"completed": map[string]any{
					"{$request.body#/url}": map[string]any{
						"post": map[string]any{"responses": map[string]any{}},
					},
					"x-callback-data": map[string]any{"x-nested": true},
				}},
			},
		}},
		"components": map[string]any{
			"schemas": map[string]any{"Payload": map[string]any{
				"properties": map[string]any{"content": map[string]any{"x-schema-extension": true}},
			}},
			"examples": map[string]any{"x-example-name": map[string]any{"summary": "valid component name"}},
			"mediaTypes": map[string]any{"application/json": map[string]any{
				"encoding": map[string]any{"first": map[string]any{
					"encoding": map[string]any{"second": map[string]any{
						"encoding": map[string]any{"third": map[string]any{}},
					}},
				}},
			}},
		},
	}

	visited := map[string]bool{}
	walker := OASWalker{visitor: func(path jp.Expr, _ map[string]any) {
		visited[path.String()] = true
	}}
	walker.Walk(document)

	for _, path := range []string{
		"$", "$.paths", "$.paths['/pets'].get",
		"$.paths['/pets'].get.responses", "$.paths['/pets'].get.responses['200']",
		"$.paths['/pets'].get.callbacks.completed",
		"$.paths['/pets'].get.callbacks.completed['{$request.body#/url}'].post",
		"$.components.mediaTypes['application/json'].encoding.first",
		"$.components.mediaTypes['application/json'].encoding.first.encoding.second",
		"$.components.mediaTypes['application/json'].encoding.first.encoding.second.encoding.third",
		"$.components.examples['x-example-name']",
	} {
		if want := jp.MustParseString(path).String(); !visited[want] {
			t.Errorf("expected %s to be visited", path)
		}
	}

	for _, path := range []string{
		"$.components.schemas.Payload.properties.content",
		"$.paths['/pets'].get.callbacks.completed['x-callback-data']",
	} {
		if want := jp.MustParseString(path).String(); visited[want] {
			t.Errorf("did not expect %s to be visited", path)
		}
	}
}

func TestPathItemVisitsEveryOperation(t *testing.T) {
	pathItem := map[string]any{}
	methods := []string{"get", "put", "post", "delete", "options", "head", "patch", "trace", "query"}
	for _, method := range methods {
		pathItem[method] = map[string]any{}
	}
	pathItem["additionalOperations"] = map[string]any{"CUSTOM": map[string]any{}}
	document := map[string]any{"paths": map[string]any{"/pets": pathItem}}

	visited := map[string]bool{}
	walker := OASWalker{visitor: func(path jp.Expr, _ map[string]any) {
		visited[path.String()] = true
	}}
	walker.Walk(document)

	for _, operation := range append(methods, "additionalOperations.CUSTOM") {
		path := jp.MustParseString("$.paths['/pets']." + operation).String()
		if !visited[path] {
			t.Errorf("expected %s operation to be visited", operation)
		}
	}
}
