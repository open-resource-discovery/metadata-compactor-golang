package oas

import (
	"encoding/json"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
)

type OASRule struct {
	Kind  string
	Value string
	Path  string
}

func (self *OASRule) Matches(expr jp.Expr, value string) bool {
	if self.Path != "" && self.Path != expr.String() {
		return false
	}

	switch self.Kind {
	case "exact":
		return self.Value == value
	case "glob":
		return value == self.Value[:len(self.Value)-2] || strings.HasPrefix(value, self.Value[:len(self.Value)-1])
	default:
		return false
	}
}

func (self *OASRule) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &self.Value)

	self.Kind = utils.Ternary(strings.HasSuffix(self.Value, "-*"), "glob", "exact")

	return err
}
