package main

import (
	"flag"
	"github.com/widmogrod/software-architecture-playground/visitor"
	"io/ioutil"
	"strings"
)

var path = flag.String("path", "-", "path to *.dpr file")
var types = flag.String("types", "", "path to *.dpr file")
var name = flag.String("name", "", "path to *.dpr file")
var packageName = flag.String("packageName", "main", "go package name")

func main() {
	flag.Parse()

	g := visitor.Generate{
		Types:       strings.Split(*types, ","),
		Name:        *name,
		PackageName: *packageName,
	}

	result, err := g.Generate()
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*path+".go", result, 0644)
	if err != nil {
		panic(err)
	}
}
