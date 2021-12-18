package datalang

import (
	"fmt"
	"strings"
)

func Fmt(ast *Ast) string {
	result := &strings.Builder{}
	for _, data := range ast.Datas {
		fmt.Fprintf(result, "data %s\n", data.Name)
		for _, constructor := range data.Body {
			if len(constructor.Value) == 0 {
				fmt.Fprintf(result, "\t| %s\n", constructor.Name)
			} else {
				fmt.Fprintf(result, "\t| %s(%s)\n", constructor.Name, strings.Join(constructor.Value, ", "))
			}
		}
		fmt.Fprintf(result, "\n")
	}

	return result.String()
}
