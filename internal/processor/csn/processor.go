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

func removeAssociationsPostprocessor(processor *CSNProcessor, document map[string]any) {
	processor.asExpression("entity").
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

func removeTypeDefinitionsPostprocessor(processor *CSNProcessor, document map[string]any) {
	processor.asExpression("type").
		Walk(document, func(path jp.Expr, _ []any) {
			path.MustRemoveOne(document)
		})
}

func resolveTypeDefinitionsPostprocessor(processor *CSNProcessor, document map[string]any) {
	processor.asExpression("entity").
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
	postprocessors []func(*CSNProcessor, map[string]any)
}

func CreateProcessor(options model.CSNOptions, rules []model.CSNRule) *CSNProcessor {
	result := &CSNProcessor{
		rules:          append([]model.CSNRule{}, rules...),
		postprocessors: make([]func(*CSNProcessor, map[string]any), 0, 3),
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

func (self *CSNProcessor) Process(document string) string {
	parsed := utils.SafeCast[map[string]any](oj.MustParseString(document))

	// Prune context,service and entity definitions
	for _, kind := range []string{"context", "service", "entity", "type"} {
		self.asExpression(kind).
			Walk(parsed, func(_ jp.Expr, nodes []any) {
				self.process(kind, nodes)
			})
	}

	for _, postprocessor := range self.postprocessors {
		postprocessor(self, parsed)
	}

	return oj.JSON(parsed)
}

func (self *CSNProcessor) prune(element map[string]any) {
	for key := range element {
		if self.shouldDelete(key) {
			delete(element, key)
		}
	}
}

func (self *CSNProcessor) shouldDelete(value string) bool {
	return (self.isAnnotation(value) || self.isPrivateProperty(value)) &&
		utils.None(self.rules, func(rule model.CSNRule) bool { return rule.Matches(value) })
}

func (self *CSNProcessor) isAnnotation(value string) bool {
	return strings.HasPrefix(value, "@")
}

func (self *CSNProcessor) process(kind string, nodes []any) {
	switch kind {
	case "type", "context", "service":
		self.prune(utils.SafeCast[map[string]any](utils.Last(nodes)))
	case "entity":
		self.prune(utils.SafeCast[map[string]any](utils.Last(nodes)))
		for _, element := range utils.SafeCast[map[string]any](utils.SafeCast[map[string]any](utils.Last(nodes))["elements"]) {
			self.prune(utils.SafeCast[map[string]any](element))
		}
	default:
		panic("unknown kind: " + kind)
	}
}

func (self *CSNProcessor) asExpression(kind string) jp.Expr {
	return jputils.Expr("$", "definitions", jputils.Eq("@.kind", kind))
}

func (self *CSNProcessor) isPrivateProperty(value string) bool {
	return strings.HasPrefix(value, "__")
}
