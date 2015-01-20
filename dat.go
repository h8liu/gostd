package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"html"
	"io"
	"io/ioutil"
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

func runeHtml(ch rune) string {
	if ch == '\t' {
		return "&nbsp;&nbsp;&nbsp;&nbsp"
	} else if ch == ' ' {
		return "&nbsp;"
	} else if ch == '\n' {
		return "<br>\n"
	}
	return html.EscapeString(string(ch))
}

func tokClass(t token.Token) string {
	switch {
	case t == token.COMMENT:
		return "comment"
	case t == token.IDENT:
		return "ident"
	case t >= token.INT && t <= token.IMAG:
		return "num"
	case t == token.CHAR || t == token.STRING:
		return "string"
	case t.IsOperator():
		return "op"
	case t.IsKeyword():
		return "keyword"
	}

	return "na"
}

func (f *file) html() []byte {
	out := new(bytes.Buffer)
	base := f.file.Base()
	bs, e := ioutil.ReadFile(f.path)
	ne(e)

	r := bytes.NewReader(bs)
	off := 0

	fmt.Fprint(out, htmlHeader)

	fmt.Fprint(out, `<div class="code">\n`)

	for _, t := range f.toks {
		if t.tok == token.EOF {
			break
		}

		toff := int(t.tok) - base

		for off < toff {
			ch, n, e := r.ReadRune()
			if e == io.EOF {
				panic("unexpected eof")
			}

			off += n

			fmt.Fprint(out, runeHtml(ch))
		}

		if off != toff {
			panic("rune not aligned")
		}

		nb := len([]byte(t.lit))

		buf := make([]byte, nb)
		_, e := r.Read(buf)
		if e != nil {
			panic("unexpected error")
		}

		str := string(buf)
		if t.lit != str {
			panic("lit unmatch")
		}

		fmt.Fprintf(out, `<span class="%s">`, tokClass(t.tok))
		for _, ch := range str {
			fmt.Fprint(out, runeHtml(ch))
		}
		fmt.Fprintf(out, `</span>`)
	}

	fmt.Fprint(out, `\n</div>\n`)

	fmt.Fprint(out, htmlFooter)

	return out.Bytes()
}
