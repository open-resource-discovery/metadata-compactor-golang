package csn

import (
	"strings"

	"github.com/huandu/go-clone"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/jputils"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
	"github.com/open-resource-discovery/metadata-compactor-golang/model"
)

func removeAssociationsPostprocessor(document map[string]any) {
	jputils.Expr("$", "definitions", jputils.Eq("@.kind", "entity")).
		Walk(document, func(_ jp.Expr, nodes []any) {
			entity := utils.SafeCast[map[string]any](utils.Last(nodes))
			elements := utils.SafeCast[map[string]any](entity["elements"])

			for name, element := range elements {
				if utils.SafeCast[map[string]any](element)["type"] == "cds.Association" {
					delete(elements, name)
				}
			}
		})
}

func removeTypeDefinitionsPostprocessor(document map[string]any) {
	jputils.Expr("$", "definitions", jputils.Eq("@.kind", "type")).
		Walk(document, func(path jp.Expr, _ []any) {
			path.MustRemoveOne(document)
		})
}

func resolveTypeDefinitionsPostprocessor(document map[string]any) {
	jputils.Expr("$", "definitions", jputils.Eq("@.kind", "entity")).
		Walk(document, func(_ jp.Expr, nodes []any) {
			entity := utils.SafeCast[map[string]any](utils.Last(nodes))

			for _, relement := range utils.SafeCast[map[string]any](entity["elements"]) {
				element := utils.SafeCast[map[string]any](relement)

				if typ := utils.SafeCast[string](element["type"]); len(typ) > 0 && !strings.HasPrefix(typ, "cds.") {
					for key, value := range utils.SafeCast[map[string]any](jputils.Expr("$", "definitions", typ).First(document)) {
						if key == "type" || (!utils.ContainsKey(element, key) && utils.OneOf(key, "on", "key", "enum", "scale", "target", "length", "notNull", "default", "precision", "cardinality")) {
							element[key] = clone.Clone(value)
						}
					}
				}
			}
		})
}

type CSNProcessor struct {
	rules          []model.CSNRule
	options        model.CSNOptions
	targets        []jp.Expr
	postprocessors []func(map[string]any)
}

func CreateProcessor(options model.CSNOptions, rules []model.CSNRule) *CSNProcessor {
	result := &CSNProcessor{
		rules:          append([]model.CSNRule{}, rules...),
		postprocessors: make([]func(map[string]any), 0, 3),
		targets: []jp.Expr{
			jputils.Expr("$", "definitions", jputils.Eq("@.kind", "context")),
			jputils.Expr("$", "definitions", jputils.Eq("@.kind", "entity")),
			jputils.Expr("$", "definitions", jputils.Eq("@.kind", "entity"), "elements", "*"),
			jputils.Expr("$", "definitions", jputils.Eq("@.kind", "service")),
			jputils.Expr("$", "definitions", jputils.Eq("@.kind", "type")),
		},
	}

	if !options.PreserveAssociations {
		result.postprocessors = append(result.postprocessors, removeAssociationsPostprocessor)
	}

	if !options.PreserveTypes {
		result.postprocessors = append(
			result.postprocessors,
			resolveTypeDefinitionsPostprocessor,
			removeTypeDefinitionsPostprocessor,
		)
	}

	return result
}

func (self *CSNProcessor) Process(document string, baselines ...string) string {
	parsed := utils.SafeCast[map[string]any](oj.MustParseString(document))
	rules := append(append([]model.CSNRule{}, self.rules...), self.extractBaselineRules(baselines...)...)

	// Prune context,service and entity definitions
	for _, target := range self.targets {
		target.Walk(parsed, func(expr jp.Expr, nodes []any) {
			self.prune(rules, expr, utils.SafeCast[map[string]any](utils.Last(nodes)))
		})
	}

	for _, postprocessor := range self.postprocessors {
		postprocessor(parsed)
	}

	return oj.JSON(parsed)
}

func (self *CSNProcessor) extractBaselineRules(baselines ...string) []model.CSNRule {
	result := make([]model.CSNRule, 0)
	extract := func(expr jp.Expr, definition map[string]any) []model.CSNRule {
		result := make([]model.CSNRule, 0)

		for key := range definition {
			if self.isAnnotation(key) || self.isPrivateProperty(key) {
				result = append(result, model.CSNRule{
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

		for _, expression := range self.targets {
			expression.Walk(parsed, func(expr jp.Expr, nodes []any) {
				result = append(result, extract(expr, utils.SafeCast[map[string]any](utils.Last(nodes)))...)
			})
		}
	}

	return result
}

func (self *CSNProcessor) prune(rules []model.CSNRule, expr jp.Expr, element map[string]any) {
	for key := range element {
		if self.shouldDelete(rules, expr, key) {
			delete(element, key)
		}
	}
}

func (self *CSNProcessor) shouldDelete(rules []model.CSNRule, expr jp.Expr, value string) bool {
	return (self.isAnnotation(value) || self.isPrivateProperty(value)) &&
		utils.None(rules, func(rule model.CSNRule) bool { return rule.Matches(expr.Child(value), value) })
}

func (self *CSNProcessor) isAnnotation(value string) bool {
	return strings.HasPrefix(value, "@")
}

func (self *CSNProcessor) isPrivateProperty(value string) bool {
	return strings.HasPrefix(value, "__")
}
