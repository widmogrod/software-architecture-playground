package elasticsearch

import (
	"fmt"
	"strings"
	"testing"
)

func TestIndexQuestion(t *testing.T) {
	schema := GenSchema(nil)
	PrintSchema(schema, 1)
	r := FlattenSchema(schema)
	for _, name := range r {
		fmt.Println(name)
	}
}

func FlattenSchema(schema Schema) []string {
	var result []string
	for _, field := range schema.Fields {
		result = append(result, FlattenTypes(field.Types, strings.ToLower(field.Name))...)
	}
	return result
}

func FlattenTypes(types Types, prev string) []string {
	var result []string
	if types.String {
		return append(result, prev)
	}

	for _, field := range types.Record {
		result = append(result, FlattenTypes(field.Types, prev+"__"+strings.ToLower(field.Name))...)
	}

	return result
}

func PrintSchema(schema Schema, depth int) {
	r := &strings.Builder{}

	fmt.Fprintf(r, "schema %s{\n", schema.Name)
	for _, field := range schema.Fields {
		fmt.Fprintf(r, "%s%s = %s\n", strings.Repeat("\t", depth), field.Name, PrintType(field.Types, depth+1))
	}
	fmt.Fprintf(r, "}\n")

	fmt.Println(r.String())
}

func PrintType(field Types, depth int) string {
	if field.String {
		return "String"
	}

	r := &strings.Builder{}
	fmt.Fprintf(r, "Record{\n")
	for _, field := range field.Record {
		fmt.Fprintf(r, "%s%s = %s\n", strings.Repeat("\t", depth), field.Name, PrintType(field.Types, depth+1))
	}
	fmt.Fprintf(r, "%s}", strings.Repeat("\t", depth-1))
	return r.String()
}
