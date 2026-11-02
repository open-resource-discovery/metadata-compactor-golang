package edmx

import (
	"fmt"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/xmltree"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

// RuleExtractor turns baseline annotations into exact target and qualifier rules.
type RuleExtractor struct{}

func (e RuleExtractor) Extract(baselines ...string) ([]edmx.EDMXRule, error) {
	var rules []edmx.EDMXRule
	for index, baseline := range baselines {
		document, err := xmltree.Convert(baseline)
		if err != nil {
			return nil, fmt.Errorf("baseline %d: %w", index+1, err)
		}
		rules = e.visit(document.Nodes(), nil, rules)
	}
	return rules, nil
}

func (e RuleExtractor) visit(nodes, parents []xmltree.Node, rules []edmx.EDMXRule) []edmx.EDMXRule {
	resolver := AnnotationTargetResolver{}
	for _, node := range nodes {
		if node.Kind() != "element" {
			continue
		}
		chain := append(parents, node)
		if localName(node) == "Annotation" {
			term, found := node.Attribute("Term")
			target, resolved := resolver.Resolve(chain)
			if found && term != "" && resolved {
				qualifier, _ := resolver.ResolveQualifier(chain)
				rules = append(rules, edmx.EDMXRule{Kind: "exact", Value: term, Target: &target, Qualifier: &qualifier})
			}
		}
		rules = e.visit(node.Nodes(), chain, rules)
	}
	return rules
}
