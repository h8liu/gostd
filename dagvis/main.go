package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ne(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	m := make(map[string][]string)

	bs, e := ioutil.ReadFile("gostd.dep")
	ne(e)

	e = json.Unmarshal(bs, &m)
	ne(e)

	nNode := len(m)
	nEdge := 0

	g := newGraph()
	for name := range m {
		g.addNode(name)
	}

	for from, tos := range m {
		nEdge += len(tos)
		for _, to := range tos {
			g.addEdge(to, from)
		}
	}

	g.buildDAG()
	g.printLevels()
	g.layout()

	bs = g.exportLayout()
	e = ioutil.WriteFile("gostd.layout", bs, 0644)
	ne(e)

	fmt.Printf("%d nodes, %d edges, %d crit edges\n",
		nNode, nEdge, g.ncrit,
	)

}
