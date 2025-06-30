package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"strings"
	"text/template"
)

var (
	methodTemplate     = template.Must(template.New("method").Parse(methodTemplateText))
	methodTemplateText = `
func (m *{{ .MockName }}) {{ .Name }}{{ .Signature }} {
	if len(m.{{ .Name }}Handlers) == 0 {
		panic("{{ .MockName}}.{{ .Name }}Handlers is empty")
	}

	f := m.{{ .Name }}Handlers[0]
	m.{{ .Name }}Handlers = m.{{ .Name }}Handlers[1:]

	return f({{ .Params }})
}`
)

type Method struct {
	MockName  string
	Name      string
	Signature string
	Call      string
	Params    string
}

func (m *Method) Render() string {
	out := bytes.Buffer{}
	methodTemplate.Execute(&out, m)

	return out.String()
}
func parseMethod(mockName string, f *ast.Field) *Method {
	if len(f.Names) < 1 {
		return nil
	}

	method := Method{
		MockName: mockName,
		Name:     f.Names[0].String(),
	}

	ft := f.Type.(*ast.FuncType)
	params := []string{}
	paramsNames := []string{}

	for i, p := range ft.Params.List {
		if len(p.Names) == 0 {
			argName := fmt.Sprintf("arg%d", i)
			params = append(params, fmt.Sprintf("%s %s", argName, getType(p.Type)))
			paramsNames = append(paramsNames, argName)
		}

		for _, name := range p.Names {
			argName := name.String()
			params = append(params, fmt.Sprintf("%s %s", argName, getType(p.Type)))
			paramsNames = append(paramsNames, argName)
		}
	}

	returns := []string{}
	if ft.Results != nil {
		for _, r := range ft.Results.List {
			if len(r.Names) == 0 {
				returns = append(returns, getType(r.Type))
			} else {
				returns = append(returns, fmt.Sprintf("%s %s", r.Names[0], getType(r.Type)))
			}
		}
	}

	method.Signature = fmt.Sprintf("(%s)%s ", strings.Join(params, ", "), formatReturns(returns))
	method.Params = strings.Join(paramsNames, ", ")

	return &method
}
