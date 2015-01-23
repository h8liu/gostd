package main

import (
	"fmt"
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

		ret[name] = outs
	}

	return ret
}
