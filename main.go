package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

func main() {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	files, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return
	}

	if len(files) > 0 {
		for _, f := range files {
			ast.Inspect(f, getPackageName)
			break
		}
	}

	for _, f := range files {
		ast.Inspect(f, inspectFile)
	}
}

func getPackageName(node ast.Node) bool {
	p, ok := node.(*ast.Package)
	if ok {
		fmt.Printf("package %s\n\n", p.Name)
		return true
	}

	return false
}

func inspectFile(node ast.Node) bool {
	t, ok := node.(*ast.TypeSpec)
	if !ok {
		return true
	}

	iType, ok := t.Type.(*ast.InterfaceType)
	if !ok {
		return true
	}

	name := fmt.Sprintf("%sMock", t.Name.String())
	methods := []Method{}
	for i := range iType.Methods.List {
		m := parseMethod(name, iType.Methods.List[i])
		if m != nil {
			methods = append(methods, *m)
		}
	}

	out := &bytes.Buffer{}
	err := mockTemplate.Execute(out, map[string]any{
		"Name":    name,
		"Methods": methods,
	})
	if err != nil {
		slog.Error("Execute mock error", "error", err, "mock", name)
		return false

	}
	fmt.Println(out.String())

	return true
}

func getType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getType(t.Elt)
	case *ast.SelectorExpr:
		return getType(t.X) + "." + getType(t.Sel)
	case *ast.StarExpr:
		return "*" + getType(t.X)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getType(t.Key), getType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	default:
		slog.Warn("Unsupported type", "type", reflect.TypeOf(t))
		return ""
	}
}

func formatReturns(returns []string) string {
	s := strings.Join(returns, ", ")
	if len(returns) <= 1 {
		return " " + s
	}

	return " (" + s + ")"
}
