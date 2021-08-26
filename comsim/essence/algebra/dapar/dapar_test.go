package dapar

import (
	"github.com/stretchr/testify/assert"
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
			assert.NoError(t, err)
			assert.Equal(t, uc.out, result)
		})
	}
}
