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
	{{typeName .Name }} struct { {{/* hack to remove empty first line*/ -}}
		{{- range .Body}}
		{{fieldName .Name }} *{{typeName .Name}}{{end}}
	}{{/* no-new-line */ -}}

{{- range .Body}}
	{{typeName .Name}} struct { {{/* hack to remove empty first line*/ -}}
		{{- range .Value -}}
		{{- if isPoli . $data}}
		{{fieldName .}} interface{} 
		{{- else}}
		{{fieldName .}} {{typeName .}}{{end}}{{end}}
	}{{end}}
)
{{ end}}
`
	render = template.Must(template.New("main").Funcs(template.FuncMap{
		"typeName":  strings.Title,
		"fieldName": strings.Title,
		"isPoli":    IsPoli,
	}).Parse(tmpl))
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
	return result.Bytes(), nil
}

func MustGenerate(result []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return result
}
