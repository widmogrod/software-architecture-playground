package datalang

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var (
	tmpl = `package {{ .PackageName }}
{{range $i, $data := .Ast.Datas}}
type (
	{{typeName .Name }} struct { {{/* no-new-line */ -}}
		{{- range $no, $constructor := .Body}}
		{{fieldName .Name }}{{$no}} *{{typeName .Name}}{{end}}
	}{{/* no-new-line */ -}}

{{- range .Body}}
	{{typeName .Name}} struct { {{/* no-new-line */ -}}
		{{- range $no, $value := .Value -}}
		{{- if isPoli . $data}}
		{{fieldName .}}{{$no}} interface{} 
		{{- else}}
		{{fieldName .}}{{$no}} *{{typeName .}}{{end}}{{end}}
	}{{end}}
)
{{ end}}
`

	tmplMake = `
{{range $i, $data := .Ast.Datas}}
{{range $no, $constructor := .Body}}
func Mk{{fieldName $constructor.Name -}}(
	{{- range $no, $value := $constructor.Value}}
		{{- if gt $no 0}}, {{end}}
		{{- if isPoli . $data -}}
		{{- argName .}}{{$no}} interface{}
		{{- else -}}
		{{- argName .}}{{$no}} *{{typeName .}}
		{{- end -}}
	{{- end -}}

) *{{typeName $data.Name}} {
	return &{{typeName $data.Name}} {
		{{fieldName $constructor.Name }}{{$no}}: &{{typeName .Name}} { {{/* no-new-line */ -}}
		{{- range $no, $value := .Value}}
			{{fieldName .}}{{$no}}: {{argName .}}{{$no}},{{end}}
		},
	}
}
{{end}}
{{end}}
`

	tmplBFS = `
{{range $i, $data := .Ast.Datas}}
func BFS_{{typeName $data.Name}}(
	{{- range $no, $constructor := .Body}}
	{{- if hasValue $data $constructor -}}
	f{{$no}} func(*{{fieldName $constructor.Name}}), 
	{{- end -}}
	{{- end -}}
	l *{{typeName $data.Name -}}
) {
	visited := map[*{{typeName $data.Name}}]bool{}
	queue := []*{{typeName $data.Name}}{l}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if visited[i] {
			continue
		}
		visited[i] = true

		{{range $no, $constructor := .Body}}
		if i.{{fieldName $constructor.Name}}{{$no}} != nil {
			{{- if hasValue $data $constructor}}
			f{{$no}}(i.{{fieldName $constructor.Name}}{{$no}})
			{{- end}}

			{{- range referenceTypes $data $constructor}}
			queue = append(queue, i.{{fieldName $constructor.Name}}{{$no}}.{{.}})
			{{- end}}

			continue
		}
		{{end}}

		panic("non-exhaustive")
	}
}
{{end}}
`

	funMap = template.FuncMap{
		"typeName":  strings.Title,
		"fieldName": strings.Title,
		"argName":   strings.ToLower,
		"isPoli":    IsPoli,
		"referenceTypes": func(d Data, c Constructor) []string {
			var result []string
			for no, s := range c.Value {
				if s == d.Name {
					result = append(result, fmt.Sprintf("%s%d", strings.Title(s), no))
				}
			}
			return result
		},
		"hasValue": func(d Data, c Constructor) bool {
			isPoli := map[string]bool{}
			for _, s := range d.Poli {
				isPoli[s] = true
			}

			for _, s := range c.Value {
				if isPoli[s] {
					return true
				}
			}

			return false
		},
	}

	render    = template.Must(template.New("main").Funcs(funMap).Parse(tmpl))
	renderMk  = template.Must(template.New("main").Funcs(funMap).Parse(tmplMake))
	renderBFS = template.Must(template.New("main").Funcs(funMap).Parse(tmplBFS))
)

type (
	genOpt struct {
		PackageName string
	}
	genData struct {
		PackageName string
		Ast         *Ast
	}
)

var (
	defaultOpt = &genOpt{PackageName: "main"}
)

type OptionBuilder = func(opt *genOpt)

func PackageName(name string) OptionBuilder {
	return func(opt *genOpt) {
		if name != "" {
			opt.PackageName = name
		}
	}
}

func IsPoli(a string, data *Data) bool {
	for _, s := range data.Poli {
		if a == s {
			return true
		}
	}

	return false
}

func GenerateGo(ast *Ast, options ...OptionBuilder) ([]byte, error) {
	opt := defaultOpt
	for _, option := range options {
		option(opt)
	}

	data := &genData{
		Ast:         ast,
		PackageName: opt.PackageName,
	}

	result := &bytes.Buffer{}
	err := render.ExecuteTemplate(result, "main", data)
	if err != nil {
		return nil, err
	}

	err = renderMk.ExecuteTemplate(result, "main", data)
	if err != nil {
		return nil, err
	}
	err = renderBFS.ExecuteTemplate(result, "main", data)
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func MustGenerate(result []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return result
}
