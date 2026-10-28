package oas

import (
	"testing"

	"github.com/ohler55/ojg/jp"
)

func TestWalkVisitsEverySupportedOASObject(t *testing.T) {
	document := map[string]any{
		"info": map[string]any{
			"contact": map[string]any{},
			"license": map[string]any{},
		},
		"externalDocs": map[string]any{},
		"tags":         []any{map[string]any{"externalDocs": map[string]any{}}},
		"servers": []any{map[string]any{
			"variables": map[string]any{"region": map[string]any{}},
		}},
		"paths": map[string]any{
			"x-path-extension": map[string]any{},
			"/pets": map[string]any{
				"servers":              []any{map[string]any{}},
				"parameters":           []any{parameterFixture()},
				"additionalOperations": map[string]any{"COPY": map[string]any{}},
				"get":                  operationFixture(),
			},
		},
		"webhooks": map[string]any{"petCreated": map[string]any{}},
		"components": map[string]any{
			"callbacks": map[string]any{"onEvent": map[string]any{
				"{$request.body#/url}": map[string]any{},
				"x-callback-extension": map[string]any{},
			}},
			"links":      map[string]any{"next": linkFixture()},
			"headers":    map[string]any{"X-Rate-Limit": headerFixture()},
			"examples":   map[string]any{"pet": map[string]any{}},
			"pathItems":  map[string]any{"pets": map[string]any{}},
			"responses":  map[string]any{"NotFound": responseFixture()},
			"parameters": map[string]any{"TraceID": parameterFixture()},
			"mediaTypes": map[string]any{"PetJSON": mediaTypeFixture()},
			"requestBodies": map[string]any{"Pet": map[string]any{
				"content": map[string]any{"application/json": mediaTypeFixture()},
			}},
			"securitySchemes": map[string]any{"OAuth": map[string]any{
				"flows": map[string]any{
					"implicit":            map[string]any{},
					"password":            map[string]any{},
					"authorizationCode":   map[string]any{},
					"clientCredentials":   map[string]any{},
					"deviceAuthorization": map[string]any{},
				},
			}},
		},
	}

	visited := make(map[string]map[string]any)
	Walk(document, func(path jp.Expr, object map[string]any) {
		visited[path.String()] = object
	})

	assertVisitedPaths(t, visited, []string{
		"$",
		"$.info", "$.info.contact", "$.info.license", "$.externalDocs",
		"$.tags[0]", "$.tags[0].externalDocs",
		"$.servers[0]", "$.servers[0].variables.region",
		"$.paths", "$.paths['/pets']", "$.paths['/pets'].servers[0]",
		"$.paths['/pets'].parameters[0]",
		"$.paths['/pets'].parameters[0].examples.sample",
		"$.paths['/pets'].parameters[0].content['application/json']",
		"$.paths['/pets'].additionalOperations.COPY",
		"$.paths['/pets'].get", "$.paths['/pets'].get.externalDocs",
		"$.paths['/pets'].get.servers[0]", "$.paths['/pets'].get.parameters[0]",
		"$.paths['/pets'].get.requestBody",
		"$.paths['/pets'].get.requestBody.content['application/json']",
		"$.paths['/pets'].get.responses",
		"$.paths['/pets'].get.responses['200']",
		"$.paths['/pets'].get.responses['200'].headers['X-Trace-ID']",
		"$.paths['/pets'].get.responses['200'].links.next",
		"$.paths['/pets'].get.responses['200'].links.next.server",
		"$.paths['/pets'].get.responses['200'].links.next.server.variables.region",
		"$.paths['/pets'].get.responses['200'].content['application/json']",
		"$.paths['/pets'].get.callbacks.onDone",
		"$.paths['/pets'].get.callbacks.onDone['{$request.body#/url}']",
		"$.webhooks.petCreated",
		"$.components", "$.components.callbacks.onEvent",
		"$.components.callbacks.onEvent['{$request.body#/url}']",
		"$.components.links.next", "$.components.headers['X-Rate-Limit']",
		"$.components.examples.pet", "$.components.pathItems.pets",
		"$.components.responses.NotFound", "$.components.parameters.TraceID",
		"$.components.mediaTypes.PetJSON", "$.components.requestBodies.Pet",
		"$.components.securitySchemes.OAuth", "$.components.securitySchemes.OAuth.flows",
		"$.components.securitySchemes.OAuth.flows.implicit",
		"$.components.securitySchemes.OAuth.flows.password",
		"$.components.securitySchemes.OAuth.flows.authorizationCode",
		"$.components.securitySchemes.OAuth.flows.clientCredentials",
		"$.components.securitySchemes.OAuth.flows.deviceAuthorization",
	})

	assertNotVisitedPaths(t, visited, []string{
		"$.paths['x-path-extension']",
		"$.paths['/pets'].get.responses['x-response-extension']",
		"$.paths['/pets'].get.callbacks.onDone['x-callback-extension']",
		"$.components.callbacks.onEvent['x-callback-extension']",
	})
}

func TestWalkHandlesMissingAndUnexpectedCollectionValues(t *testing.T) {
	document := map[string]any{
		"tags":     "not-an-array",
		"servers":  nil,
		"webhooks": []any{},
		"paths":    "not-an-object",
		"components": map[string]any{"mediaTypes": map[string]any{
			"empty": map[string]any{},
		}},
	}

	visited := make(map[string]map[string]any)
	Walk(document, func(path jp.Expr, object map[string]any) { visited[path.String()] = object })

	assertVisitedPaths(t, visited, []string{"$", "$.info", "$.paths", "$.components.mediaTypes.empty"})
	if visited[jp.MustParseString("$.paths").String()] != nil {
		t.Fatal("expected a non-object paths value to be exposed to the visitor as nil")
	}
}

func operationFixture() map[string]any {
	return map[string]any{
		"externalDocs": map[string]any{},
		"servers":      []any{map[string]any{}},
		"parameters":   []any{parameterFixture()},
		"requestBody": map[string]any{
			"content": map[string]any{"application/json": mediaTypeFixture()},
		},
		"responses": map[string]any{
			"200":                  responseFixture(),
			"x-response-extension": map[string]any{},
		},
		"callbacks": map[string]any{"onDone": map[string]any{
			"{$request.body#/url}": map[string]any{},
			"x-callback-extension": map[string]any{},
		}},
	}
}

func parameterFixture() map[string]any {
	return map[string]any{
		"examples": map[string]any{"sample": map[string]any{}},
		"content":  map[string]any{"application/json": mediaTypeFixture()},
	}
}

func responseFixture() map[string]any {
	return map[string]any{
		"links":   map[string]any{"next": linkFixture()},
		"headers": map[string]any{"X-Trace-ID": headerFixture()},
		"content": map[string]any{"application/json": mediaTypeFixture()},
	}
}

func headerFixture() map[string]any {
	return map[string]any{
		"examples": map[string]any{"sample": map[string]any{}},
		"content":  map[string]any{"application/json": mediaTypeFixture()},
	}
}

func linkFixture() map[string]any {
	return map[string]any{"server": map[string]any{
		"variables": map[string]any{"region": map[string]any{}},
	}}
}

func mediaTypeFixture() map[string]any {
	return map[string]any{
		"examples":     map[string]any{"sample": map[string]any{}},
		"itemEncoding": map[string]any{},
		"encoding": map[string]any{"property": map[string]any{
			"headers":  map[string]any{"X-Encoded": map[string]any{}},
			"encoding": map[string]any{"nested": map[string]any{}},
		}},
		"prefixEncoding": []any{map[string]any{}},
	}
}

func assertVisitedPaths(t *testing.T, visited map[string]map[string]any, paths []string) {
	t.Helper()
	for _, path := range paths {
		canonical := jp.MustParseString(path).String()
		if _, ok := visited[canonical]; !ok {
			t.Errorf("expected %s to be visited", path)
		}
	}
}

func assertNotVisitedPaths(t *testing.T, visited map[string]map[string]any, paths []string) {
	t.Helper()
	for _, path := range paths {
		canonical := jp.MustParseString(path).String()
		if _, ok := visited[canonical]; ok {
			t.Errorf("did not expect %s to be visited", path)
		}
	}
}
