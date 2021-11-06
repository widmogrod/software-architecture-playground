package dapar

import (
	"bytes"
	"fmt"
	"strings"
)

type Config struct {
	PackageName string
}

func Generate(ast *Ast, c *Config) ([]byte, error) {
	isAliasToDataType := map[string]string{}
	result := &bytes.Buffer{}
	tags := &bytes.Buffer{}

	fmt.Fprintf(result, "// GENERATED do not edit!\n")
	fmt.Fprintf(result, "package %s\n", c.PackageName)

	for _, dt := range ast.DataTypes {
		fmt.Fprintf(result, "\ntype (\n")
		fmt.Fprintf(result, "\t%s interface {\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "\t	_union%s()\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "\t}\n")
		for _, dc := range dt.Sum {
			if dc.Alias != nil {
				isAliasToDataType[*dc.Alias] = dt.Name
				continue
			}

			result.Write(renderFlatType(dc))

			fmt.Fprintf(tags, "func (_ %s) _union%s() {}\n", strings.Title(dc.Name), strings.Title(dt.Name))
			if dataTypeAlias, ok := isAliasToDataType[dt.Name]; ok {
				fmt.Fprintf(tags, "func (_ %s) _union%s() {} // Alias\n", strings.Title(dc.Name), strings.Title(dataTypeAlias))
			}
		}
		fmt.Fprintf(result, ")\n")
		tags.WriteTo(result)

		if dt.Sum != nil {
			fmt.Fprintf(result, "\n")
			result.Write(renderVisitor(dt))
			fmt.Fprintf(result, "\n")
			result.Write(renderVisitorMap(dt))
		}
	}

	return result.Bytes(), nil
}

func typeName(t *Typ) string {
	if t.Record != nil {
		result := strings.Builder{}
		for _, record := range t.Record {
			result.WriteString(strings.Title(record.Key))
		}
		return result.String()
	}

	return strings.Title(t.Name)
}

func typeSuffix(path []*Typ) string {
	result := strings.Builder{}
	for i := range path {
		result.WriteString(typeName(path[i]))
	}
	return result.String()
}

func isLeaf(t *Typ) bool {
	if t.Name != "" {
		return true
	}
	if t.List != nil && isLeaf(t.List) {
		return true
	}

	return false
}

func getLeafName(t *Typ) string {
	if t.Name != "" {
		return strings.Title(t.Name)
	}

	return "[]" + getLeafName(t.List)
}

func getNonLeafName(t *Typ, dcName string, path []*Typ) string {
	suffix := ""
	if len(path) > 0 {
		suffix = "Literal"
		if t.Record != nil {
			suffix = "Record"
		} else if t.List != nil {
			suffix = "List"
		} else if t.Tuple != nil {
			suffix = "Tuple"
		}
	}

	return dcName + typeSuffix(path) + suffix
}

func renderFlatType(dc DataConstructor) []byte {
	result := bytes.NewBuffer(nil)
	dcName := strings.Title(dc.Name)

	if dc.Args == nil {
		fmt.Fprintf(result, "\t%s struct {}\n", dcName)
		return result.Bytes()
	}

	BreathFirstSearch(dc.Args, func(t *Typ, path []*Typ) {
		// Top level list rendering
		if t.List != nil && len(path) == 0 {
			if isLeaf(t) {
				fmt.Fprintf(result, "\t%s []%s // leaf\n", getNonLeafName(t, dcName, path), getLeafName(t.List))
			} else {
				fmt.Fprintf(result, "\t%s []%s // non-leaf\n", getNonLeafName(t, dcName, path), getNonLeafName(t.List, dcName, append(path, t)))
			}
		}

		if t.Record != nil {
			fmt.Fprintf(result, "\t%s struct {\n", getNonLeafName(t, dcName, path))
			for _, record := range t.Record {
				if isLeaf(record.Value) {
					fmt.Fprintf(result, "\t\t%s %s\n", strings.Title(record.Key), getLeafName(record.Value))
				} else {
					fmt.Fprintf(result, "\t\t%s %s\n", strings.Title(record.Key), getNonLeafName(record.Value, dcName, append(path, t)))
				}
			}
			fmt.Fprintf(result, "\t}\n")
		}

		if t.Tuple != nil {
			fmt.Fprintf(result, "\t%s struct {\n", getNonLeafName(t, dcName, path))
			for i := range t.Tuple {
				tuple := t.Tuple[i]
				if isLeaf(tuple) {
					fmt.Fprintf(result, "\t\tT%d %s\n", i+1, getLeafName(tuple))
				} else {
					fmt.Fprintf(result, "\t\tT%d %s\n", i+1, getNonLeafName(tuple, dcName, append(path, t)))
				}
			}
			fmt.Fprintf(result, "\t}\n")
		}
	})

	return result.Bytes()
}

type VisitorFunc = func(typ *Typ, path []*Typ)

func BreathFirstSearch(typ *Typ, f VisitorFunc) {
	visited := make(map[*Typ]bool)
	parent := make(map[*Typ]*Typ)

	for queue := []*Typ{typ}; len(queue) > 0; {
		typ := queue[0]
		queue = queue[1:]

		if _, ok := visited[typ]; ok {
			continue
		}

		path := make([]*Typ, 0)
		p := typ
		for {
			if p = parent[p]; p != nil {
				path = append([]*Typ{p}, path...)
			} else {
				break
			}
		}

		f(typ, path)
		visited[typ] = true

		if typ.Record != nil {
			for _, record := range typ.Record {
				parent[record.Value] = typ
				queue = append(queue, record.Value)
			}
		} else if typ.Tuple != nil {
			for i := range typ.Tuple {
				tuple := typ.Tuple[i]
				parent[tuple] = typ
				queue = append(queue, tuple)
			}
		} else if typ.List != nil {
			parent[typ.List] = typ
			queue = append(queue, typ.List)
		}
	}
}

func renderVisitor(dt DataType) []byte {
	result := bytes.NewBuffer(nil)
	dtName := strings.Title(dt.Name)

	fmt.Fprintf(result, "type %sVisitor interface {\n", dtName)
	for _, dc := range dt.Sum {
		dcName := strings.Title(dc.Name)
		fmt.Fprintf(result, "\tVisit%s(x %s) interface{}\n", dcName, dcName)
	}
	fmt.Fprintf(result, "}\n")

	return result.Bytes()
}

func renderVisitorMap(dt DataType) []byte {
	result := bytes.NewBuffer(nil)
	dtName := strings.Title(dt.Name)

	fmt.Fprintf(result, "func Map%s(value %s, v %sVisitor) interface{} {\n", dtName, dtName, dtName)
	fmt.Fprintf(result, "\tswitch x := value.(type) {\n")
	for _, dc := range dt.Sum {
		dcName := strings.Title(dc.Name)
		fmt.Fprintf(result, "\tcase %s:\n", dcName)
		fmt.Fprintf(result, "\t\treturn v.Visit%s(x)\n", dcName)
	}
	fmt.Fprintf(result, "\tdefault:\n")
	fmt.Fprintf(result, "\t\tpanic(`unknown type`)\n")
	fmt.Fprintf(result, "\t}\n")
	fmt.Fprintf(result, "}\n")

	return result.Bytes()
}
