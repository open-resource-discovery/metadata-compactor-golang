package oas

import (
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/oas"
)

type OASProcessor struct {
	rules []oas.OASRule
}

func CreateProcessor(rules []oas.OASRule) *OASProcessor {
	return &OASProcessor{rules: append([]oas.OASRule{}, rules...)}
}

func (self *OASProcessor) Process(document string, baselines ...string) string {
	parsed := utils.SafeCast[map[string]any](oj.MustParseString(document))
	rules := append(append([]oas.OASRule{}, self.rules...), self.extractBaselineRules(baselines...)...)

	Walk(parsed, func(expr jp.Expr, object map[string]any) {
		self.prune(rules, expr, object)
	})

	return oj.JSON(parsed)
}

func (self *OASProcessor) extractBaselineRules(baselines ...string) []oas.OASRule {
	result := make([]oas.OASRule, 0)
	extract := func(expr jp.Expr, definition map[string]any) []oas.OASRule {
		result := make([]oas.OASRule, 0)

		for key := range definition {
			if self.isSpecificationExtensions(key) {
				result = append(result, oas.OASRule{
					Value: key,
					Kind:  "exact",
					Path:  expr.Child(key).String(),
				})
			}
		}

		return result
	}

	for _, baseline := range baselines {
		parsed := utils.SafeCast[map[string]any](oj.MustParseString(baseline))

		Walk(parsed, func(expr jp.Expr, object map[string]any) {
			result = append(result, extract(expr, object)...)
		})
	}

	return result
}

func (self *OASProcessor) prune(rules []oas.OASRule, expr jp.Expr, element map[string]any) {
	for key := range element {
		if self.shouldDelete(rules, expr, key) {
			delete(element, key)
		}
	}
}

func (self *OASProcessor) shouldDelete(rules []oas.OASRule, expr jp.Expr, value string) bool {
	return self.isSpecificationExtensions(value) &&
		utils.None(rules, func(rule oas.OASRule) bool { return rule.Matches(expr.Child(value), value) })
}

func (self *OASProcessor) isSpecificationExtensions(value string) bool {
	return strings.HasPrefix(value, "x-")
}
