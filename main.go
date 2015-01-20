package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
)

func listPackages() []string {
	c := &build.Default
	hold := c.GOPATH
	c.GOPATH = ""
	pkgs := buildutil.AllPackages(c)
	c.GOPATH = hold

	var ret []string
	for _, p := range pkgs {
		if strings.HasPrefix(p, "cmd/") || p == "cmd" {
			continue
		}
		if p == "runtime/cgo" {
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
		for _, f := range p.files {
			fmt.Printf("   %s (%s)\n", f.name, f.path)
			f.parseToks()
		}
	}
}
