//go:build unit

package edmx_test

import (
	"encoding/json"
	"testing"

	"github.com/open-resource-discovery/metadata-compactor-golang/model"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

func TestEDMXRulesetJSON(t *testing.T) {
	var ruleset model.Ruleset
	if err := json.Unmarshal([]byte(`{"edmx":{"preserve":["Core.Description","Common.*"]}}`), &ruleset); err != nil {
		t.Fatal(err)
	}
	rules := ruleset.EDMX.Rules
	if len(rules) != 2 || rules[0].Kind != "exact" || rules[1].Kind != "glob" {
		t.Fatalf("unexpected rules: %#v", rules)
	}
	for _, rule := range rules {
		if rule.Target != nil || rule.Qualifier != nil {
			t.Fatal("configured term rules must match every target and qualifier")
		}
	}
	for _, content := range []string{`{}`, `{"edmx":{}}`, `{"edmx":null}`, `{"edmx":false}`, `{"edmx":{"preserve":[]}}`} {
		if err := json.Unmarshal([]byte(content), &ruleset); err != nil {
			t.Fatalf("%s: %v", content, err)
		}
		// Fresh decoding also covers an omitted format block.
		var fresh model.Ruleset
		if err := json.Unmarshal([]byte(content), &fresh); err != nil || len(fresh.EDMX.Rules) != 0 {
			t.Fatalf("%s: %#v, %v", content, fresh, err)
		}
	}
	for _, content := range []string{`{"preserve":null}`, `{"preserve":"One"}`, `{"preserve":[null]}`, `{"preserve":[1]}`, `{"preserve":[{}]}`} {
		var rules edmx.EDMXRuleset
		if err := json.Unmarshal([]byte(content), &rules); err == nil {
			t.Fatalf("accepted invalid rules: %s", content)
		}
	}
}
