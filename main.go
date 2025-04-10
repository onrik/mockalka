package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
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

	fmt.Println(`import (
	"github.com/stretchr/testify/mock"
)`)

	for _, f := range files {
		ast.Inspect(f, inspectFile)
	}
}

func getPackageName(node ast.Node) bool {
	packge, ok := node.(*ast.Package)
	if ok {
		fmt.Println("package", packge.Name, "\n")
		return true
	}

	return false
}

func inspectFile(node ast.Node) bool {
	t, ok := node.(*ast.TypeSpec)
	if !ok {
		return true
	}

	i, ok := t.Type.(*ast.InterfaceType)
	if !ok {
		return true
	}
	out := []string{}
	name := fmt.Sprintf("%sMock", t.Name.String())
	out = append(out, fmt.Sprintf("type %s struct {", name), "\tmock.Mock", "}", "")

	for _, f := range i.Methods.List {
		if len(f.Names) < 1 {
			continue
		}
		funcName := f.Names[0]

		ft := f.Type.(*ast.FuncType)
		called := []string{}
		params := []string{}

		for i, p := range ft.Params.List {
			if len(p.Names) == 0 {
				argName := fmt.Sprintf("arg%d", i)
				params = append(params, fmt.Sprintf("%s %s", argName, getType(p.Type)))
				called = append(called, argName)
			}

			for _, name := range p.Names {
				argName := name.String()
				params = append(params, fmt.Sprintf("%s %s", argName, getType(p.Type)))
				called = append(called, argName)
			}
		}

		args := []string{}
		returns := []string{}
		if ft.Results != nil {
			for i, r := range ft.Results.List {
				if len(r.Names) == 0 {
					returns = append(returns, getType(r.Type))
				} else {
					returns = append(returns, fmt.Sprintf("%s %s", r.Names[0], getType(r.Type)))
				}
				t := getType(r.Type)
				arg := fmt.Sprintf("args.Get(%d)", i)
				if t == "error" {
					arg = fmt.Sprintf("args.Error(%d)", i)
				} else {
					arg = fmt.Sprintf("%s.(%s)", arg, t)
				}
				args = append(args, arg)
			}
		}

		out = append(
			out,
			fmt.Sprintf("func (m *%s) %s(%s)%s {", name, funcName, strings.Join(params, ", "), formatReturns(returns)),
		)

		if len(args) > 0 {
			out = append(
				out,
				fmt.Sprintf("\targs := m.Called(%s)", strings.Join(called, ", ")),
				fmt.Sprintf("\treturn %s", strings.Join(args, ", ")),
			)
		}

		out = append(
			out,
			"}",
			"",
		)
	}

	fmt.Println()
	fmt.Println(strings.Join(out, "\n"))
	fmt.Println()
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
		log.Println("Unsupported type:", reflect.TypeOf(t))
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
