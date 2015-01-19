package main

import (
	"go/token"
)

type tok struct {
	pos token.Pos   // token position in the file set
	tok token.Token // token type
	lit string
}

type file struct {
	file *token.File
	name string
	path string

	toks []*tok
}

type pkg struct {
	path  string
	files []*file
}
