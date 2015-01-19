package main

import (
	"fmt"
	"go/build"
	"strings"

	"golang.org/x/tools/go/buildutil"
)

func packages() []string {
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

func main() {
	pkgs := packages()
	for _, p := range pkgs {
		fmt.Println(p)
	}
}
