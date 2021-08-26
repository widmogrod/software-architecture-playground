package dapar

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	useCases := map[string]struct {
		in  []byte
		out *Ast
		err error
	}{
		"should return empty AST when there is nothing to parse": {
			in:  []byte(``),
			out: &Ast{},
		},
		"should return error when there is incomplete input": {
			in:  []byte(`abc`),
			out: &Ast{},
			err: errors.New("cannot make sence with: &dapar.Ident{found:[]uint8{0x61, 0x62, 0x63}}; what left: "),
		},
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
						{Name: "in", Args: []Typ{
							{Name: "a"},
						}},
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
						{Name: "in", Args: []Typ{
							{Name: "lit"},
							{Name: "book"},
						}},
					}},
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			result, err := Parse(uc.in)
			if uc.err != nil || err != nil {
				assert.Equal(t, uc.err, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, uc.out, result)
		})
	}
}

func TestGenerate(t *testing.T) {
	c := &Config{PackageName: "_assets"}
	useCases := map[string]struct {
		c    *Config
		ast  *Ast
		file string
	}{
		"should generate nothing on input = ''": {
			c:    c,
			ast:  &Ast{},
			file: "_assets/empty_gen.go",
		},
		"should generate nothing on input = 'maybe = Just(a) | Nothing`'": {
			c:    c,
			ast:  MustParse([]byte(`maybe = Just(maybe) | Nothing`)),
			file: "_assets/maybe_gen.go",
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			result, err := Generate(uc.ast, uc.c)
			assert.NoError(t, err)
			assert.FileExists(t, uc.file)
			expected, err := ioutil.ReadFile(uc.file)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(result))
		})
	}
}
