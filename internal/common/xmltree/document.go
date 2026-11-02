package xmltree

type Document map[string]any

func (d Document) Nodes() []Node {
	nodes, _ := d["nodes"].([]Node)
	return nodes
}
