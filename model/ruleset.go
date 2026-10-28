package model

import (
	"time"

	"github.com/open-resource-discovery/metadata-compactor-golang/model/csn"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/oas"
)

type Ruleset struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`

	CSN csn.CSNRuleset `json:"csn"`
	OAS oas.OASRuleset `json:"oas"`
}
