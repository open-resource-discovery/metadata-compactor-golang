package edmx

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type EDMXRuleset struct {
	Rules []EDMXRule `json:"preserve"`
}

// UnmarshalJSON mirrors Node's optional EDMX block and validates preserve lists.
func (r *EDMXRuleset) UnmarshalJSON(data []byte) error {
	*r = EDMXRuleset{}
	data = bytes.TrimSpace(data)
	// Node treats a missing or non-object format block as an empty ruleset.
	if len(data) == 0 || data[0] != '{' {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	preserve, found := fields["preserve"]
	if !found {
		return nil
	}
	preserve = bytes.TrimSpace(preserve)
	if len(preserve) == 0 || preserve[0] != '[' {
		return fmt.Errorf("edmx.preserve must be an array of strings")
	}
	return json.Unmarshal(preserve, &r.Rules)
}
