package main

import (
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
)

func main() {
	invoker.StartDockerRuntime(func(input invoker.FunctionInput) invoker.FunctionOutput {
		return "Hello %s, from Docker!"
	})
}
