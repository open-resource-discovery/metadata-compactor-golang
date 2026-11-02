package edmx

import (
	"fmt"
	"slices"

	"github.com/huandu/go-clone"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/xml2json"
	"github.com/open-resource-discovery/metadata-compactor-golang/model/edmx"
)

type EDMXProcessor struct {
	rules     []edmx.EDMXRule
	extractor RuleExtractor
}

// CreateProcessor takes an independent snapshot of the configured rules.
func CreateProcessor(rules []edmx.EDMXRule) *EDMXProcessor {
	return &EDMXProcessor{rules: clone.Clone(rules).([]edmx.EDMXRule), extractor: RuleExtractor{}}
}

func (p *EDMXProcessor) Process(content string, baselines ...string) (string, error) {
	document, err := xml2json.Convert(content)
	if err != nil {
		return "", fmt.Errorf("EDMX input: %w", err)
	}
	baselineRules, err := p.extractor.Extract(baselines...)
	if err != nil {
		return "", err
	}
	document["nodes"] = p.prune(document.Nodes(), nil, slices.Concat(p.rules, baselineRules))

	return document.ToXML(), nil
}

func (p *EDMXProcessor) prune(nodes, parents []xml2json.Node, rules []edmx.EDMXRule) []xml2json.Node {
	result := make([]xml2json.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Type() == "element" {
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

func shouldPreserve(nodes []xml2json.Node, rules []edmx.EDMXRule) bool {
	term := nodes[len(nodes)-1].Attribute("Term")
	if term == "" {
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
