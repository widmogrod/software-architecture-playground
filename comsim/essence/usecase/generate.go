// build ignore

package main

import (
	"flag"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/dapar"
	"io/ioutil"
)

var path = flag.String("path", "-", "path to *.dpr file")
var packageName = flag.String("packageName", "main", "go package name")

//go:generate go run generate.go -path "data/runtime.dpr" -packageName data
///go:generate go run generate.go -path "lang/lang.dpr" -packageName lang

func main() {
	flag.Parse()

	path := *path
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	ast, err := dapar.Parse(data)
	if err != nil {
		panic(err)
	}
	result, err := dapar.Generate(ast, &dapar.Config{
		PackageName: *packageName,
	})
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(path+".go", result, 0644)
	if err != nil {
		panic(err)
	}
}
