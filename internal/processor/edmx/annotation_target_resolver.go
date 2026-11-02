package edmx

import (
	"strings"

	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/xmltree"
)

// AnnotationTargetResolver follows CSDL 4.01 target identities used by Node PR #5.
type AnnotationTargetResolver struct{}

func (r AnnotationTargetResolver) Resolve(nodes []xmltree.Node) (string, bool) {
	if len(nodes) == 0 || localName(nodes[len(nodes)-1]) != "Annotation" {
		return "", false
	}
	return r.resolveElement(nodes[:len(nodes)-1])
}

// ResolveQualifier inherits block qualifiers only for direct child annotations.
func (r AnnotationTargetResolver) ResolveQualifier(nodes []xmltree.Node) (string, bool) {
	if len(nodes) == 0 || localName(nodes[len(nodes)-1]) != "Annotation" {
		return "", false
	}
	annotation := nodes[len(nodes)-1]
	if qualifier, found := annotation.Attribute("Qualifier"); found {
		return qualifier, true
	}
	if len(nodes) >= 2 {
		parent := nodes[len(nodes)-2]
		if localName(parent) == "Annotations" {
			return parent.Attribute("Qualifier")
		}
	}
	return "", false
}

func (r AnnotationTargetResolver) resolveElement(nodes []xmltree.Node) (string, bool) {
	if len(nodes) == 0 {
		return "", false
	}
	element := nodes[len(nodes)-1]
	parents := nodes[:len(nodes)-1]
	var parent xmltree.Node
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	}
	switch name := localName(element); name {
	case "Annotations":
		target, _ := element.Attribute("Target")
		return nonempty(target)
	case "Annotation":
		term, _ := element.Attribute("Term")
		if _, found := nonempty(term); !found {
			return "", false
		}
		qualifier, _ := r.ResolveQualifier(nodes)
		suffix := "@" + term
		if qualifier != "" {
			suffix += "#" + qualifier
		}
		target, _ := r.resolveElement(parents)
		return appendTarget(target, suffix)
	case "Schema":
		namespace, _ := element.Attribute("Namespace")
		return nonempty(namespace)
	case "Action", "Function":
		if localName(parent) != "Schema" {
			return "", false
		}
		namespace, _ := parent.Attribute("Namespace")
		return operationTarget(namespace, element)
	case "EnumType", "EntityType", "ComplexType", "EntityContainer", "TypeDefinition", "Term":
		if localName(parent) != "Schema" {
			return "", false
		}
		namespace, _ := parent.Attribute("Namespace")
		value, _ := element.Attribute("Name")
		return qualify(namespace, value)
	case "Parameter", "ReturnType":
		if localName(parent) != "Action" && localName(parent) != "Function" {
			return "", false
		}
		target, _ := r.resolveElement(parents)
		suffix, _ := element.Attribute("Name")
		if name == "ReturnType" {
			suffix = "$ReturnType"
		}
		return appendTarget(target, suffix)
	case "Member", "Property", "NavigationProperty", "EntitySet", "Singleton", "ActionImport", "FunctionImport":
		target, _ := r.resolveElement(parents)
		value, _ := element.Attribute("Name")
		return appendTarget(target, value)
	default:
		// Expressions and other non-addressable elements do not inherit a model target.
		return "", false
	}
}

func operationTarget(namespace string, operation xmltree.Node) (string, bool) {
	value, _ := operation.Attribute("Name")
	name, found := qualify(namespace, value)
	if !found {
		return "", false
	}
	isAction := localName(operation) == "Action"
	binding, _ := operation.Attribute("IsBound")
	isBound := binding == "true" || binding == "1"
	if isAction && !isBound {
		return name + "()", true
	}
	var types []string
	for _, child := range operation.Nodes() {
		if child.Kind() != "element" || localName(child) != "Parameter" {
			continue
		}
		value, _ := child.Attribute("Type")
		typ, found := nonempty(value)
		if !found {
			return "", false
		}
		if isAction {
			return name + "(" + typ + ")", true
		}
		types = append(types, typ)
	}
	if isBound && len(types) == 0 {
		return "", false
	}
	return name + "(" + strings.Join(types, ",") + ")", true
}

func nonempty(value string) (string, bool) {
	if isWhitespaceOnly(value) {
		return "", false
	}
	return value, true
}

func appendTarget(target, name string) (string, bool) {
	if _, found := nonempty(target); !found {
		return "", false
	}
	if _, found := nonempty(name); !found {
		return "", false
	}
	return target + "/" + name, true
}

func qualify(namespace, name string) (string, bool) {
	if _, found := nonempty(namespace); !found {
		return "", false
	}
	if _, found := nonempty(name); !found {
		return "", false
	}
	return namespace + "." + name, true
}

func localName(node xmltree.Node) string {
	name := node.Name()
	return name[strings.LastIndex(name, ":")+1:]
}
