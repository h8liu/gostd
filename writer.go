package main

import (
	"bytes"
	"fmt"
	"go/token"
	"html"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type writer struct {
	outRoot string // html root
	outPath string // web path

	pkgs  map[string]*pkg
	fset  *token.FileSet
	files map[int]*file

	quiet     bool
	linkToken bool
}

func (w *writer) writePkgs() {
	e := os.MkdirAll(filepath.Join(w.outRoot, w.outPath), 0700)
	ne(e)

	for _, p := range w.pkgs {
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

			bs := w.html(p, f)
			pout := filepath.Join(outPath, f.name+".html")
			e := ioutil.WriteFile(pout, bs, 0700)
			ne(e)
		}

		bs := w.pkgHtml(p)
		pout := filepath.Join(outPath, "index.html")
		e = ioutil.WriteFile(pout, bs, 0700)
		ne(e)
	}

	bs := w.homepage()
	pout := filepath.Join(w.outRoot, w.outPath, "index.html")
	e = ioutil.WriteFile(pout, bs, 0700)
	ne(e)
}

func (w *writer) htmlPkgNavi(p *pkg, out io.Writer) {
	var files []string

	ppath := p.path
	for _, f := range p.files {
		files = append(files, f.name)
	}

	if strings.HasSuffix(ppath, "_test") {
		ppath2 := strings.TrimSuffix(ppath, "_test")
		p2, found := w.pkgs[ppath2]
		if found {
			for _, f := range p2.files {
				files = append(files, f.name)
			}
		}
	} else {
		p2, found := w.pkgs[ppath+"_test"]
		if found {
			for _, f := range p2.files {
				files = append(files, f.name)
			}
		}
	}

	fmt.Fprintln(out, `<div class="pkgnavi">`)

	fmt.Fprintf(out, `<h1><a href="%s">Go Standard Library</a></h1>`+"\n",
		html.EscapeString(path.Join("/", w.outPath, "index.html")),
	)

	fmt.Fprintf(out, `<h2>%s</h2>`+"\n",
		html.EscapeString(p.savePath),
	)

	if len(files) > 0 {
		sort.Strings(files)
		fmt.Fprintln(out, `<ul>`)
		for _, f := range files {
			fmt.Fprintf(out, `<li><a href="%s">%s</a></li>`+"\n",
				html.EscapeString(w.href(path.Join(p.savePath, f))),
				html.EscapeString(f),
			)
		}

		fmt.Fprintln(out, `</ul>`)
	}

	fmt.Fprintln(out, `</div>`)
}

func (w *writer) htmlBody(f *file, out io.Writer) {
	base := f.file.Base()
	size := f.file.Size()
	bs, e := ioutil.ReadFile(f.path)
	ne(e)
	r := bytes.NewReader(bs)

	off := 0

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
				dest := w.href(defFile.savePath)
				dest += fmt.Sprintf("#%d", int(defPos))

				fmt.Fprintf(out, `<a href="%s">`,
					html.EscapeString(dest),
				)
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

}

func (w *writer) html(p *pkg, f *file) []byte {
	out := new(bytes.Buffer)

	fmt.Fprint(out, htmlHeader)
	w.htmlPkgNavi(p, out)
	w.htmlBody(f, out)
	fmt.Fprint(out, htmlFooter)

	return out.Bytes()
}

func (w *writer) pkgHtml(p *pkg) []byte {
	out := new(bytes.Buffer)
	fmt.Fprint(out, htmlHeader)
	w.htmlPkgNavi(p, out)
	fmt.Fprint(out, htmlFooter)

	return out.Bytes()
}

func (w *writer) homepage() []byte {
	out := new(bytes.Buffer)
	fmt.Fprint(out, htmlHeader)

	fmt.Fprintf(out, `<h1><a href="%s">Go Standard Library</a></h1>`+"\n",
		html.EscapeString(path.Join("/", w.outPath, "index.html")),
	)

	if len(w.pkgs) == 0 {
		panic("no package")
	}

	var pkgs []string
	hits := make(map[string]struct{})

	for _, p := range w.pkgs {
		if _, found := hits[p.savePath]; found {
			continue
		}
		pkgs = append(pkgs, p.savePath)
		hits[p.savePath] = struct{}{}
	}

	sort.Strings(pkgs)

	fmt.Fprintln(out, `<div class="pkgs"><ul class="pkgs">`)
	for _, p := range pkgs {
		fmt.Fprintf(out, `<li><a href="%s">%s</a></li>`+"\n",
			html.EscapeString(path.Join("/", w.outPath, p)),
			html.EscapeString(p),
		)
	}
	fmt.Fprintln(out, `</ul></div>`)

	fmt.Fprint(out, htmlFooter)
	return out.Bytes()
}

func (w *writer) href(savePath string) string {
	return path.Join("/", w.outPath, savePath+".html")
}
