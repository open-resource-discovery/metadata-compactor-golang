package edmx

import (
	"fmt"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/xmltree"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

type EDMXProcessor struct {
	rules []edmx.EDMXRule
}

// CreateProcessor takes an independent snapshot of the configured rules.
func CreateProcessor(rules []edmx.EDMXRule) *EDMXProcessor {
	copied := make([]edmx.EDMXRule, len(rules))
	for index, rule := range rules {
		if rule.Target != nil {
			target := *rule.Target
			rule.Target = &target
		}
		if rule.Qualifier != nil {
			qualifier := *rule.Qualifier
			rule.Qualifier = &qualifier
		}
		copied[index] = rule
	}
	return &EDMXProcessor{rules: copied}
}

func (p *EDMXProcessor) Process(content string, baselines ...string) (string, error) {
	document, err := xmltree.Convert(content)
	if err != nil {
		return "", fmt.Errorf("EDMX input: %w", err)
	}
	baselineRules, err := (RuleExtractor{}).Extract(baselines...)
	if err != nil {
		return "", err
	}
	rules := make([]edmx.EDMXRule, 0, len(p.rules)+len(baselineRules))
	rules = append(rules, p.rules...)
	rules = append(rules, baselineRules...)
	document["nodes"] = p.prune(document.Nodes(), nil, rules)

	// Node emits XML declarations and DOCTYPEs before other document nodes.
	var declarations, directives, nodes []xmltree.Node
	for _, node := range document.Nodes() {
		switch node.Kind() {
		case "declaration":
			declarations = append(declarations, node)
		case "directive":
			directives = append(directives, node)
		default:
			nodes = append(nodes, node)
		}
	}
	document["nodes"] = append(append(declarations, directives...), nodes...)
	return document.ToXML()
}

func (p *EDMXProcessor) prune(nodes, parents []xmltree.Node, rules []edmx.EDMXRule) []xmltree.Node {
	result := make([]xmltree.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind() == "comment" || (node.Kind() == "text" && isWhitespaceOnly(node.Value())) {
			continue
		}
		if node.Kind() == "element" {
			chain := append(parents, node)
			if localName(node) == "Annotation" && !shouldPreserve(chain, rules) {
				continue
			}
			node["nodes"] = p.prune(node.Nodes(), chain, rules)
		}
		result = append(result, node)
	}
	return result
}

func shouldPreserve(nodes []xmltree.Node, rules []edmx.EDMXRule) bool {
	term, found := nodes[len(nodes)-1].Attribute("Term")
	if !found {
		return false
	}
	resolver := AnnotationTargetResolver{}
	target, _ := resolver.Resolve(nodes)
	qualifier, _ := resolver.ResolveQualifier(nodes)
	for _, rule := range rules {
		if rule.Matches(target, term, qualifier) {
			return true
		}
	}
	return false
}
