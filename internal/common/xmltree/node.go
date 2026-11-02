package xmltree

type Node map[string]any

func (n Node) Kind() string {
	return n.stringField("kind")
}

func (n Node) Name() string {
	return n.stringField("name")
}

func (n Node) LocalName() string {
	return n.stringField("localName")
}

func (n Node) NamespaceURI() string {
	return n.stringField("namespaceURI")
}

func (n Node) Value() string {
	return n.stringField("value")
}

func (n Node) Nodes() []Node {
	nodes, _ := n["nodes"].([]Node)
	return nodes
}

func (n Node) Attributes() map[string]string {
	attributes, _ := n["attributes"].(map[string]string)
	return attributes
}

func (n Node) Attribute(name string) (string, bool) {
	value, found := n.Attributes()[name]
	return value, found
}

func (n Node) AttributeNamespace(name string) string {
	namespaces, _ := n["attributeNamespaces"].(map[string]string)
	return namespaces[name]
}

func (n Node) stringField(name string) string {
	value, _ := n[name].(string)
	return value
}
