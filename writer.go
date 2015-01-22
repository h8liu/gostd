package main

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

type writer struct {
	outRoot string // html root
	outPath string // web path

	pkgs  []*pkg
	fset  *token.FileSet
	files map[int]*file

	quiet     bool
	linkToken bool
}

func (w *writer) writePkgs(pkgs []*pkg) {
	for _, p := range pkgs {
		if !w.quiet {
			fmt.Printf("[%s]\n", p.path)
		}

		outPath := filepath.Join(w.outRoot, w.outPath, p.savePath)

		e := os.MkdirAll(outPath, 0700)
		ne(e)

		for _, f := range p.files {
			if !w.quiet {
				fmt.Printf("   %s (%s)\n", f.name, f.path)
			}

			f.parseToks()

			bs := w.html(f)
			pout := filepath.Join(outPath, f.name+".html")
			e := ioutil.WriteFile(pout, bs, 0700)
			ne(e)
		}
	}
}

func (w *writer) html(f *file) []byte {
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

	writeRune := func(ch rune) {
		fmt.Fprint(out, runeHtml(ch))
		if ch == '\n' {
			line++
			fmt.Fprintf(out, lineHeader(line))
		}
	}

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
			writeRune(ch)
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

		var defPos token.Pos
		if w.linkToken {
			defPos = token.Pos(f.refs[int(t.pos)])
			if defPos != 0 {
				defFile := w.files[w.fset.File(defPos).Base()]
				if defFile == nil {
					panic("def file missing")
				}
				dest := path.Join("/", w.outPath, defFile.savePath+".html")
				dest += fmt.Sprintf("#%d", int(defPos))

				fmt.Fprintf(out, `<a href="%s">`, dest)
			}
		}

		for _, ch := range str {
			writeRune(ch)
		}

		if w.linkToken && defPos != 0 {
			fmt.Fprintf(out, `</a>`)
		}

		fmt.Fprintf(out, `</span>`)

		off += nb
	}

	fmt.Fprint(out, "\n</div>\n")

	fmt.Fprint(out, htmlFooter)

	return out.Bytes()
}
