package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"html"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

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

var builtinTypes = []string{
	"int",
	"uint",
	"int32",
	"uint32",
	"int16",
	"uint16",
	"int8",
	"uint8",
	"byte",
	"rune",
	"error",
	"string",
	"float32",
	"float64",
	"complex64",
	"complex128",
	"uintptr",
	"bool",
	"map",
	"true",
	"false",
}

var builtinFuncs = []string{
	"len",
	"cap",
	"close",
	"complex",
	"delete",
	"imag",
	"panic",
	"print",
	"println",
	"real",
	"recover",
	"make",
	"append",
	"new",
	"copy",
}

var builtinFuncMap = makeMap(builtinFuncs)
var builtinTypeMap = makeMap(builtinTypes)

func makeMap(lst []string) map[string]struct{} {
	ret := make(map[string]struct{})
	for _, s := range lst {
		ret[s] = struct{}{}
	}

	return ret
}

func tokClass(t token.Token, lit string) string {
	switch {
	case t == token.COMMENT:
		return "comment"
	case t == token.IDENT:
		if _, found := builtinFuncMap[lit]; found {
			return "builtinfunc"
		} else if _, found := builtinTypeMap[lit]; found {
			return "builtintype"
		}
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

func lineHeader(line int) string {
	if line == 1 || line%5 == 0 {
		return fmt.Sprintf(`<span class="lineno">%d</span>`, line)
	}
	return fmt.Sprintf(`<span class="lineno"></span>`)
}

func (f *file) html(fset *token.FileSet, files map[int]*file) []byte {
	out := new(bytes.Buffer)
	base := f.file.Base()
	size := f.file.Size()

	bs, e := ioutil.ReadFile(f.path)
	ne(e)

	r := bytes.NewReader(bs)
	off := 0

	fmt.Fprint(out, htmlHeader)

	fmt.Fprint(out, `<div class="code">`+"\n")
	line := 1
	fmt.Fprint(out, lineHeader(line))

	for _, t := range f.toks {
		if t.tok == token.EOF {
			break
		}
		if t.lit == "\n" && t.tok == token.SEMICOLON {
			// implicit semicolon
			continue
		}

		toff := int(t.pos) - base
		if toff < 0 || toff >= size {
			log.Printf("invalid toff=%d", toff)
			panic("invalid toff")
		}

		for off < toff {
			ch, n, e := r.ReadRune()
			if e == io.EOF {
				panic("unexpected eof")
			}

			off += n

			fmt.Fprint(out, runeHtml(ch))
			if ch == '\n' {
				line++
				fmt.Fprint(out, lineHeader(line))
			}
		}

		if off != toff {
			log.Printf("off=%d toff=%d", off, toff)
			panic("rune not aligned")
		}

		lit := t.lit
		if t.tok.IsOperator() {
			lit = t.tok.String()
		}

		nb := len([]byte(lit))

		buf := make([]byte, nb)
		_, e := r.Read(buf)
		if e != nil {
			panic("unexpected error")
		}

		str := string(buf)

		if lit != str {
			log.Printf("t=%s lit=%q str=%q", t.tok.String(), lit, str)
			panic("lit unmatch")
		}

		fmt.Fprintf(out, `<span class="%s" id="%d">`,
			tokClass(t.tok, lit), int(t.pos),
		)

		defPos := token.Pos(f.refs[int(t.pos)])
		if defPos != 0 {
			defFile := files[fset.File(defPos).Base()]
			if defFile == nil {
				panic("def file missing")
			}
			dest := filepath.Join("/", defFile.savePath+".html")
			dest += fmt.Sprintf("#%d", int(defPos))

			fmt.Fprintf(out, `<a href="%s">`, dest)
		}

		for _, ch := range str {
			fmt.Fprint(out, runeHtml(ch))
			if ch == '\n' {
				line++
				fmt.Fprint(out, lineHeader(line))
			}
		}

		if defPos != 0 {
			fmt.Fprintf(out, `</a>`)
		}

		fmt.Fprintf(out, `</span>`)

		off += nb
	}

	fmt.Fprint(out, "\n</div>\n")

	fmt.Fprint(out, htmlFooter)

	return out.Bytes()
}
