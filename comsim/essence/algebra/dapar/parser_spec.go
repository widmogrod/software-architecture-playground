package dapar

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ParserFunc = func([]byte) (*Ast, error)

type UseCases = map[string]Spec

type Spec struct {
	in  []byte
	out *Ast
	err error
}

func SpecRunner(t *testing.T, parse ParserFunc, useCases UseCases) {
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			result, err := parse(uc.in)
			if uc.err != nil || err != nil {
				assert.Equal(t, uc.err, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, uc.out, result)

			// For better visual debugging
			g, _ := Generate(uc.out, &Config{PackageName: "g"})
			h, _ := Generate(result, &Config{PackageName: "g"})
			assert.Equal(t, string(g), string(h))
		})
	}
}

var (
	AdvanceSpec = map[string]Spec{
		"should return AST with a few data types": {
			in: []byte(`data = in; 
other = out`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "in"},
					}},
					{Name: "other", Sum: []DataConstructor{
						{Name: "out"},
					}},
				},
			},
		},
		"should return AST with data constructor that accept list": {
			in: []byte(`data = many([in])`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "many", Args: ref(tuple(
							list("in"),
						))},
					}},
				},
			},
		},
		"should return AST with data constructor that accept list 2": {
			in: []byte(`data = many([in]) | more [to]`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "many", Args: ref(tuple(
							list("in"),
						))},
						{Name: "more", Args: ref(
							list("to"),
						)},
					}},
				},
			},
		},
	}
	TrivialSpec = map[string]Spec{
		"should return empty AST when there is nothing to parse": {
			in:  []byte(``),
			out: &Ast{},
		},
		//"should return error when there is incomplete input": {
		//	in:  []byte(`abc`),
		//	out: &Ast{},
		//	err: errors.New("cannot make sence with: &dapar.Ident{found:[]uint8{0x61, 0x62, 0x63}}; what left: "),
		//},
		"should return AST when there is simple type": {
			in: []byte(`data = in`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "in"},
					}},
				},
			},
		},
		"should return AST when there is sum type": {
			in: []byte(`data = in | out`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "in", Args: nil},
						{Name: "out", Args: nil},
					}},
				},
			},
		},
		"should return AST when there is constructor type": {
			in: []byte(`data = in(a) | out`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "in", Args: ref(tuple(typ("a")))},
						{Name: "out"},
					}},
				},
			},
		},
		"should return AST when there is constructor complex types": {
			in: []byte(`data = in(lit, book)`),
			out: &Ast{
				DataTypes: []DataType{
					{Name: "data", Sum: []DataConstructor{
						{Name: "in", Args: ref(tuple(
							typ("lit"),
							typ("book"),
						))},
					}},
				},
			},
		},
	}
)

func ref(t Typ) *Typ {
	return &t
}

func typ(n string) Typ {
	return Typ{Name: n}
}
func tuple(xs ...Typ) Typ {
	ys := make([]*Typ, len(xs))
	for i := range xs {
		ys[i] = &xs[i]
	}

	return Typ{Tuple: ys}
}

func list(n string) Typ {
	return Typ{List: &Typ{Name: n}}
}
