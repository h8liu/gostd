package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

func ne(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	rev := flag.Bool("rev", false, "reverse order")
	out := flag.String("o", "gostd.layout", "output file")
	flag.Parse()

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
			if !*rev {
				g.addEdge(to, from)
			} else {
				g.addEdge(from, to)
			}
		}
	}

	g.buildDAG()
	g.printLevels()
	g.layout()

	bs = g.exportLayout()
	e = ioutil.WriteFile(*out, bs, 0644)
	ne(e)

	fmt.Printf("%d nodes, %d edges, %d crit edges\n",
		nNode, nEdge, g.ncrit,
	)

}
