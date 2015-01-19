package main

import (
	"fmt"
	"go/build"
	// "go/token"
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

	for _, pinfo := range prog.AllPackages {
		if len(pinfo.Files) == 0 {
			continue
		}
		fmt.Printf("[%s]\n", pinfo.Pkg.Path())
		for _, f := range pinfo.Files {
			pos := fset.Position(f.Package)
			base := filepath.Base(pos.Filename)
			if !strings.HasSuffix(base, ".go") {
				continue
			}

			fmt.Printf("   %s\n", base)
		}
	}
}
