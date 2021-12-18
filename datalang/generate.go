package datalang

import (
	"bytes"
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
// package {{ .PackageName }}

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

	funMap = template.FuncMap{
		"typeName":  strings.Title,
		"fieldName": strings.Title,
		"argName":   strings.ToLower,
		"isPoli":    IsPoli,
	}

	render   = template.Must(template.New("main").Funcs(funMap).Parse(tmpl))
	renderMk = template.Must(template.New("main").Funcs(funMap).Parse(tmplMake))
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
	return result.Bytes(), nil
}

func MustGenerate(result []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return result
}
