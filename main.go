package main

import (
	"go/build"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/refactor/lexical"
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

func makePkgs() (map[string]*pkg, *token.FileSet, map[int]*file) {
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

	ps := make(map[string]*pkg)
	files := make(map[int]*file)

	for tpkg, pinfo := range prog.AllPackages {
		if len(pinfo.Files) == 0 {
			continue
		}

		p := new(pkg)
		p.path = pinfo.Pkg.Path()
		p.savePath = strings.TrimSuffix(p.path, "_test")
		p.tpkg = tpkg
		p.fileMap = make(map[int]*file)

		for _, f := range pinfo.Files {
			pos := fset.Position(f.Package)
			pfile := fset.File(f.Package)
			fname := pos.Filename
			base := filepath.Base(pos.Filename)
			if !strings.HasSuffix(base, ".go") {
				continue
			}

			newFile := &file{
				file:     pfile,
				name:     base,
				path:     fname,
				savePath: filepath.Join(p.savePath, base),
				refs:     make(map[int]int),
			}

			p.files = append(p.files, newFile)
			p.fileMap[pfile.Base()] = newFile
			files[pfile.Base()] = newFile
		}

		if len(p.files) == 0 {
			continue
		}

		ps[p.path] = p
	}

	for tpkg, pinfo := range prog.AllPackages {
		lex := lexical.Structure(fset, tpkg, &pinfo.Info, pinfo.Files)
		for obj, refs := range lex.Refs {
			defPos := obj.Pos()
			for _, ref := range refs {
				refPos := ref.Id.NamePos
				refFile := fset.File(refPos)
				f := files[refFile.Base()]
				if f == nil {
					panic("file not found")
				}

				f.refs[int(refPos)] = int(defPos)
			}
		}
	}

	return ps, fset, files
}

func main() {
	ps, fset, files := makePkgs()
	w := &writer{
		outRoot: "www",
		outPath: "gostd",

		pkgs:  ps,
		fset:  fset,
		files: files,
	}

	w.writePkgs()
}
