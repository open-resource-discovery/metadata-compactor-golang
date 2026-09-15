package model

type CSNRuleset struct {
	Options CSNOptions `json:"options"`
	Rules   []CSNRule  `json:"preserve"`
}
