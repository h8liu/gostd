package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
)

func listPackages() []string {
	c := &build.Default
	c.CgoEnabled = false

	hold := c.GOPATH
	c.GOPATH = ""
	pkgs := buildutil.AllPackages(c)
	c.GOPATH = hold

	var ret []string
	for _, p := range pkgs {
		if strings.HasPrefix(p, "cmd/") || p == "cmd" {
			continue
		}
		ret = append(ret, p)
	}
	return ret
}

func loadPackages(pkgs []string) (*loader.Program, error) {
	conf := loader.Config{
		Build:         &build.Default,
		SourceImports: true,
	}

	for _, p := range pkgs {
		e := conf.ImportWithTests(p)
		if e != nil {
			return nil, e
		}
	}

	return conf.Load()
}

func makePkgs() []*pkg {
	pkgs := listPackages()
	pkgs = append(pkgs, 
		"lonnie.io/e8vm",
		"lonnie.io/e8vm/arch8",
		"lonnie.io/e8vm/dasm8",
		"lonnie.io/e8vm/lex8",
		"lonnie.io/e8vm/link8",
		"lonnie.io/e8vm/asm8",
	)
	prog, e := loadPackages(pkgs)
	ne(e)

	fset := prog.Fset

	var ps []*pkg

	for _, pinfo := range prog.AllPackages {
		if len(pinfo.Files) == 0 {
			continue
		}

		p := new(pkg)
		p.path = pinfo.Pkg.Path()

		for _, f := range pinfo.Files {
			pos := fset.Position(f.Package)
			pfile := fset.File(f.Package)
			fname := pos.Filename
			base := filepath.Base(pos.Filename)
			if !strings.HasSuffix(base, ".go") {
				continue
			}

			p.files = append(p.files, &file{
				file: pfile,
				name: base,
				path: fname,
			})
		}

		if len(p.files) == 0 {
			continue
		}

		ps = append(ps, p)
	}

	return ps
}

func main() {
	ps := makePkgs()
	for _, p := range ps {
		fmt.Printf("[%s]\n", p.path)
		outPath := filepath.Join("www", p.path)
		outPath = strings.TrimSuffix(outPath, "_test")
		e := os.MkdirAll(outPath, 0700)
		ne(e)

		for _, f := range p.files {
			fmt.Printf("   %s (%s)\n", f.name, f.path)
			f.parseToks()

			bs := f.html()
			pout := filepath.Join(outPath, f.name+".html")
			e := ioutil.WriteFile(pout, bs, 0700)
			ne(e)
		}
	}
}
