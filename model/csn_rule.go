package model

import (
	"encoding/json"
	"strings"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
)

type CSNRule struct {
	Kind  string
	Value string
}

func (self *CSNRule) Matches(value string) bool {
	switch self.Kind {
	case "exact":
		return self.Value == value
	case "glob":
		return strings.HasPrefix(value, self.Value[:len(self.Value)-1]) // value is '@x.y.z' and self.Value is '@x.y.*'
	default:
		return false
	}
}

func (self *CSNRule) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &self.Value)
	self.Kind = utils.Ternary(strings.HasSuffix(self.Value, ".*"), "glob", "exact")

	return err
}
