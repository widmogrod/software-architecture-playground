package main

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
)

func main() {
	invoker.StartDockerRuntime(func(input invoker.FunctionInput) invoker.FunctionOutput {
		return fmt.Sprintf("Hello %s, from Docker!", input)
	})
}
