package main

import (
	"flag"
	"github.com/widmogrod/software-architecture-playground/datalang"
	"io/ioutil"
)

var path = flag.String("path", "-", "path to *.dpr file")
var packageName = flag.String("packageName", "main", "go package name")

func main() {
	flag.Parse()
	path := *path
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	ast, err := datalang.Parse(data)
	if err != nil {
		panic(err)
	}
	result, err := datalang.GenerateGo(ast, datalang.PackageName(*packageName))
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(path+".go", result, 0644)
	if err != nil {
		panic(err)
	}
}
