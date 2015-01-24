package main

import (
	"fmt"
)

type Node struct {
	name string
	ins  map[string]*Node
	outs map[string]*Node

	allIns  map[string]*Node
	allOuts map[string]*Node

	critIns  map[string]*Node
	critOuts map[string]*Node

	level int
	x, y  float64

	inLvlMax  int
	inLvlMin  int
	outLvlMax int
	outLvlMin int

	nhit int
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

	levels [][]*Node

	ncrit int
}

func newGraph() *Graph {
	ret := new(Graph)
	ret.nodes = make(map[string]*Node)
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

func (g *Graph) reset() {
	for _, node := range g.nodes {
		node.level = 0
		node.x = 0
		node.y = 0
		node.nhit = 0

		node.allIns = make(map[string]*Node)
		node.allOuts = make(map[string]*Node)
		node.critIns = make(map[string]*Node)
		node.critOuts = make(map[string]*Node)
	}

	g.levels = nil
}

func (g *Graph) buildLevels() {
	var cur []*Node

	for _, node := range g.nodes {
		if len(node.ins) == 0 {
			cur = append(cur, node)
		}
	}

	nNode := 0

	for len(cur) > 0 {
		for _, node := range cur {
			for _, out := range node.outs {
				for _, in := range node.allIns {
					out.allIns[in.name] = in
					in.allOuts[out.name] = out
				}

				out.allIns[node.name] = node
				node.allOuts[out.name] = out
			}

			node.level = len(g.levels)
		}

		g.levels = append(g.levels, cur)
		nNode += len(cur)

		var next []*Node
		for _, node := range cur {
			for _, out := range node.outs {
				out.nhit++
				if out.nhit == len(out.ins) {
					next = append(next, out)
				}
			}
		}

		cur = next
	}

	if nNode != len(g.nodes) {
		panic("graph is not a dag")
	}
}

func (g *Graph) buildCrits() {
	g.ncrit = 0
	for _, node := range g.nodes {
		for _, out := range node.outs {
			isCrit := true
			for _, via := range node.allOuts {
				if via == out {
					continue
				}

				if via.allOuts[out.name] != nil {
					// node -> via
					// via -> out
					// node -> out is not critical

					isCrit = false
					break
				}
			}

			if isCrit {
				node.critOuts[out.name] = out
				out.critIns[node.name] = node
				g.ncrit++
			}
		}
	}
}

func (g *Graph) buildDAG() {
	g.reset()
	g.buildLevels()
	g.buildCrits()
}

func (g *Graph) printLevels() {
	for i, lvl := range g.levels {
		fmt.Printf("%d:", i)

		for _, node := range lvl {
			fmt.Printf(" %s", node.name)
		}

		fmt.Println()
	}
}
