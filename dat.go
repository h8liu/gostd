package main

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"

	"golang.org/x/tools/go/types"
)

type tok struct {
	pos token.Pos   // token position in the file set
	tok token.Token // token type
	lit string
}

type file struct {
	file     *token.File
	name     string
	path     string
	savePath string

	toks []*tok

	refs map[int]int
}

type pkg struct {
	path     string
	savePath string
	tpkg     *types.Package
	files    []*file
	fileMap  map[int]*file
}

func (f *file) parseToks() {
	bs, e := ioutil.ReadFile(f.path)
	ne(e)

	onError := func(pos token.Position, msg string) {
		fmt.Printf("%s: %s\n", pos, msg)
	}
	s := new(scanner.Scanner)
	s.Init(f.file, bs, onError, scanner.ScanComments)

	f.toks = nil

	for {
		p, t, l := s.Scan()
		f.toks = append(f.toks, &tok{pos: p, tok: t, lit: l})
		if t == token.EOF {
			break
		}
	}
}
