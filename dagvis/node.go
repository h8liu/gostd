package main

type Node struct {
	name string
	ins  map[string]*Node
	outs map[string]*Node

	allIns  map[string]*Node
	allOuts map[string]*Node

	critIns  map[string]*Node
	critOuts map[string]*Node

	level int
	x, y float64
}

func newNode(n string) *Node {
	ret := new(Node)
	ret.name = n
	ret.ins = make(map[string]*Node)
	ret.outs = make(map[string]*Node)
	
	return ret
}

type Graph struct {
	nodes map[string]*Node
}

func newGraph() *Graph {
	ret := new(Graph)
	return ret
}

func (g *Graph) getNode(n string) *Node {
	return g.nodes[n]
}

func (g *Graph) addNode(n string) *Node {
	if g.getNode(n) != nil {
		panic("node already exists")
	}
	
	ret := newNode(n)
	g.nodes[n] = ret
	return ret
}

func (g *Graph) addEdge(from, to string) {
	nodeFrom := g.getNode(from)
	nodeTo := g.getNode(to)

	if nodeFrom == nil || nodeTo == nil {
		panic("not not exists")
	}

	nodeFrom.outs[to] = nodeTo
	nodeTo.ins[from] = nodeFrom
}
