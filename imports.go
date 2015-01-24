package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/types"
)

type importNode struct {
	name      string
	outs      map[string]*importNode
	ins       map[string]*importNode
	nimported int
}

func newImportNode(name string) *importNode {
	ret := new(importNode)
	ret.name = name
	ret.ins = make(map[string]*importNode)
	ret.outs = make(map[string]*importNode)

	return ret
}

type importMap struct {
	nodes map[string]*importNode
}

func (m *importMap) registerNode(name string) {
	if _, found := m.nodes[name]; found {
		return
	}

	node := newImportNode(name)
	m.nodes[name] = node
}

func newImportMap(pkgs map[string]*pkg) *importMap {
	m := new(importMap)
	m.nodes = make(map[string]*importNode)

	for _, p := range pkgs {
		if strings.HasSuffix(p.path, "_test") {
			continue // skip test packages
		}
		m.registerNode(p.savePath)
	}

	path := func(p *types.Package) string {
		return p.Path()
	}

	for _, p := range pkgs {
		if strings.HasSuffix(p.path, "_test") {
			continue // skip test packages
		}

		importing := path(p.tpkg)
		nodeImporting := m.nodes[importing]

		imports := p.tpkg.Imports()
		for _, im := range imports {
			imported := path(im)
			if imported == importing {
				continue
			}
			m.registerNode(imported)
			nodeImported := m.nodes[imported]

			nodeImporting.outs[imported] = nodeImported
			nodeImported.ins[importing] = nodeImporting
		}
	}

	return m
}

func (m *importMap) levels() [][]*importNode {
	for _, node := range m.nodes {
		node.nimported = 0
	}

	var ret [][]*importNode
	var cur []*importNode
	total := 0

	for _, node := range m.nodes {
		if len(node.ins) == 0 {
			cur = append(cur, node)
		}
	}

	for len(cur) > 0 {
		ret = append(ret, cur)
		total += len(cur)

		/*
			fmt.Printf("%d:", len(ret))
			for _, node := range cur {
				fmt.Printf(" %s", node.name)
			}
			fmt.Println()
		*/

		var next []*importNode
		for _, node := range cur {
			for _, child := range node.outs {
				child.nimported++
				// fmt.Printf("hitting %s %d/%d\n", child.name,
				//	child.nimported, len(child.imports))
				if child.nimported == len(child.ins) {
					next = append(next, child)
				}
			}
		}

		cur = next
	}

	if total != len(m.nodes) {
		for _, node := range m.nodes {
			if node.nimported < len(node.ins) {
				fmt.Println(node.name)
			}

			for _, imp := range node.ins {
				if imp.nimported < len(imp.ins) {
					fmt.Println("    ", imp.name)
				}
			}
		}
		panic("have circles")
	}

	nlevel := len(ret)
	rev := make([][]*importNode, nlevel)
	for i, nodes := range ret {
		rev[nlevel-1-i] = nodes
	}

	return rev
}

func (m *importMap) export() map[string][]string {
	ret := make(map[string][]string)

	for name, node := range m.nodes {
		outs := make([]string, 0, 10)
		for p := range node.outs {
			outs = append(outs, p)
		}

		sort.Strings(outs)
		ret[name] = outs
	}

	return ret
}

func (m *importMap) export2() []byte {
	type Label struct {
		Label string `json:"label"`
	}

	type Node struct {
		Id    string `json:"id"`
		Value Label  `json:"value"`
	}

	type Link struct {
		U     string `json:"u"`
		V     string `json:"v"`
		Value Label  `json:"value"`
	}

	type Data struct {
		Name  string  `json:"name"`
		Nodes []*Node `json:"nodes"`
		Links []*Link `json:"links"`
	}

	ret := m.export()

	var froms []string
	for from := range ret {
		froms = append(froms, from)
	}
	sort.Strings(froms)

	obj := new(Data)
	obj.Name = "gostd"

	for _, from := range froms {
		obj.Nodes = append(obj.Nodes, &Node{
			Id:    from,
			Value: Label{from},
		})

		tos := ret[from]
		for _, to := range tos {
			obj.Links = append(obj.Links, &Link{
				V:     from,
				U:     to,
				Value: Label{""},
			})
		}
	}

	bs, e := json.MarshalIndent(obj, "", "   ")
	ne(e)
	return bs
}
