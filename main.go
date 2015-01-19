package main

import (
	"fmt"
	"go/build"
	"go/token"
	"log"
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

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	pkgs := listPackages()
	prog, e := loadPackages(pkgs)
	ne(e)

	fset := prog.Fset

	var fnames []string
	prog.Fset.Iterate(func(f *token.File) bool {
		fnames = append(fnames, f.Name())
		return true
	})
	for _, f := range fnames {
		fmt.Println(f)
	}

	for _, pinfo := range prog.AllPackages {
		if len(pinfo.Files) == 0 {
			continue
		}
		fmt.Printf("[%s]\n", pinfo.Pkg.Name())
		for _, f := range pinfo.Files {
			pos := fset.Position(f.Package)
			base := filepath.Base(pos.Filename)
			if !strings.HasSuffix(base, ".go") {
				panic(base)
			}

			fmt.Printf("   %s\n", base)
		}
	}
}
