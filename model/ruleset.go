package model

import "time"

type Ruleset struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`

	CSN CSNRuleset `json:"csn"`
}
