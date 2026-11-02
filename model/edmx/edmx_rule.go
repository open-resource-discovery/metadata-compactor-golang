package edmx

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EDMXRule matches a literal Term, optionally restricted to a target and qualifier.
type EDMXRule struct {
	Kind      string
	Value     string
	Target    *string // nil matches every target.
	Qualifier *string // nil matches all qualifiers; an empty string matches only unqualified annotations.
}

func (r EDMXRule) Matches(target, name, qualifier string) bool {
	if r.Target != nil && *r.Target != target {
		return false
	}
	if r.Qualifier != nil && *r.Qualifier != qualifier {
		return false
	}
	switch r.Kind {
	case "exact":
		return r.Value == name
	case "glob":
		return strings.HasSuffix(r.Value, ".*") && strings.HasPrefix(name, strings.TrimSuffix(r.Value, "*"))
	default:
		return false
	}
}

func (r *EDMXRule) UnmarshalJSON(data []byte) error {
	var value *string
	if err := json.Unmarshal(data, &value); err != nil || value == nil {
		return fmt.Errorf("edmx.preserve must be an array of strings")
	}
	*r = EDMXRule{Kind: "exact", Value: *value}
	if strings.HasSuffix(*value, ".*") {
		r.Kind = "glob"
	}
	return nil
}
