package main

import (
	"fmt"
	"go/token"
	"html"
)

func runeHtml(ch rune) string {
	if ch == '\t' {
		return "&nbsp;&nbsp;&nbsp;&nbsp;"
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
	"nil",
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
