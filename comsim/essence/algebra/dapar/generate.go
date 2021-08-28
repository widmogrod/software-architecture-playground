package dapar

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Config struct {
	PackageName string
}

func Generate(ast *Ast, c *Config) ([]byte, error) {
	isAliasToDataType := map[string]string{}
	result := &bytes.Buffer{}

	fmt.Fprintf(result, "// GENERATED do not edit!\n")
	fmt.Fprintf(result, "package %s\n", c.PackageName)

	for _, dt := range ast.DataTypes {
		fmt.Fprintf(result, "\ntype %s interface {\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "	_union%s()\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "}\n")
		for _, dc := range dt.Sum {
			if dc.Alias != nil {
				isAliasToDataType[*dc.Alias] = dt.Name
				continue
			}
			fmt.Fprintf(result, "\ntype %s ", strings.Title(dc.Name))
			gentypes(result, dc.Args, 0)
			fmt.Fprintf(result, "\nfunc (_ %s) _union%s() {}\n", strings.Title(dc.Name), strings.Title(dt.Name))
			if dataTypeAlias, ok := isAliasToDataType[dt.Name]; ok {
				fmt.Fprintf(result, "\nfunc (_ %s) _union%s() {} // Alias\n", strings.Title(dc.Name), strings.Title(dataTypeAlias))
			}
		}
	}

	return result.Bytes(), nil
}

func gentypes(result io.Writer, t *Typ, depth int) {
	if t == nil {
		fmt.Fprintf(result, "struct {}\n")
		return
	}

	if t.Record != nil {
		fmt.Fprintf(result, "struct {\n")
		for _, record := range t.Record {
			result.Write(bytes.Repeat([]byte("\t"), depth+1))
			fmt.Fprintf(result, "%s ", strings.Title(record.Key))
			gentypes(result, record.Value, depth+1)
		}
		result.Write(bytes.Repeat([]byte("\t"), depth))
		fmt.Fprintf(result, "}\n")
		return
	}

	if t.Tuple != nil {
		fmt.Fprintf(result, "struct {\n")
		for i, t := range t.Tuple {
			result.Write(bytes.Repeat([]byte("\t"), depth+1))
			fmt.Fprintf(result, "T%d ", i+1)
			gentypes(result, t, depth+1)
		}
		result.Write(bytes.Repeat([]byte("\t"), depth))
		fmt.Fprintf(result, "}\n")
		return
	}

	if t.List != nil {
		fmt.Fprintf(result, "[]")
		gentypes(result, t.List, depth)
		return
	}

	fmt.Fprintf(result, "%s\n", strings.Title(t.Name))
}
