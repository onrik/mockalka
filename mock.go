package main

import "text/template"

var (
	mockTemplate     = template.Must(template.New("mock").Parse(mockTemplateText))
	mockTemplateText = `
type {{ .Name }} struct { {{ range $method := .Methods }}
	{{ $method.Name }}Handlers []func {{ $method.Signature }}{{end}}
}

func (m *{{ .Name }}) AssertExpectations(t *testing.T) { {{ range $method := .Methods }}
	if len(m.{{ $method.Name }}Handlers) > 0 {
		t.Errorf("Not all {{ $.Name }}.{{ $method.Name }}Handlers has been called")
	}{{ end }}
}

{{ range $method := .Methods }}
	{{ $method.Render }}
{{end}}
`
)
