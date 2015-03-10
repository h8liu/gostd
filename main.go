package main

import (
	"encoding/json"
	"flag"
	"go/build"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/refactor/lexical"

	"fmt"
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

func loadPackages(pkgs []string, withTests bool) (*loader.Program, error) {
	conf := loader.Config{
		Build:         &build.Default,
	}

	for _, p := range pkgs {
		if withTests {
			conf.ImportWithTests(p)
		} else {
			conf.Import(p)
		}
	}

	return conf.Load()
}

func makePkgs(withTests bool) (map[string]*pkg, *token.FileSet, map[int]*file) {
	pkgs := listPackages()
	/*
		pkgs = append(pkgs,
			"lonnie.io/e8vm",
			"lonnie.io/e8vm/arch8",
			"lonnie.io/e8vm/dasm8",
			"lonnie.io/e8vm/lex8",
			"lonnie.io/e8vm/link8",
			"lonnie.io/e8vm/asm8",
		)
	*/

	prog, e := loadPackages(pkgs, withTests)
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
		for use, obj := range pinfo.Uses {
			defPos := obj.Pos()
			refPos := use.NamePos
			refFile := fset.File(refPos)
			f := files[refFile.Base()]
			f.refs[int(refPos)] = int(defPos)
		}

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

				old := f.refs[int(refPos)]
				if old > 0 && old != int(defPos) {
					panic("result different from types")
				}
				f.refs[int(refPos)] = int(defPos)
			}
		}
	}

	return ps, fset, files
}

func main() {
	doHtmls := flag.Bool("html", false, "create html files")
	doLevels := flag.Bool("lvl", false, "create level mapping")
	depOut := flag.String("dep", "gostd.dep", "output of dependency file")
	flag.Parse()

	if *doHtmls {
		ps, fset, files := makePkgs(true)
		w := &writer{
			outRoot: "www",
			outPath: "gostd",

			pkgs:      ps,
			fset:      fset,
			files:     files,
			linkToken: true,
		}

		w.writePkgs()
	}

	if *doLevels {
		ps, _, _ := makePkgs(false)
		m := newImportMap(ps)

		/*
			for _, node := range m.nodes {
				fmt.Println(node.name)
				for _, imp := range node.imports {
					fmt.Println("   ", imp.name)
				}
			}
		*/

		lvls := m.levels()
		for lvl, nodes := range lvls {
			fmt.Printf("%d:", lvl+1)
			for _, node := range nodes {
				fmt.Printf(" %s", node.name)
			}
			fmt.Printf("\n")
		}

		if *depOut != "" {
			deps := m.export()
			bs, e := json.MarshalIndent(deps, "", "    ")
			ne(e)

			// bs := m.export2()
			e = ioutil.WriteFile(*depOut, bs, 0644)
			ne(e)
		}
	}
}
