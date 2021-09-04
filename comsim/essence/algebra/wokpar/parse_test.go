package wokpar

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//type ParserFunc = func([]byte) (*Ast, error)

type UseCases = map[string]Spec

type Spec struct {
	in  []byte
	out *Ast
	err error
}

//func strPtr(s string) *string {
//	return &s
//}

func TestParser_TrivialSpec(t *testing.T) {
	useCases := UseCases{
		"should return AST with a few data types": {
			in: []byte(`flow start(input) {
	_ = ReserveAvailability(input.Id)
	if and(exists(input.do), eq(a.ID, ddd.d.d.d)) {
		b = ProcessPayment(input)
		return(b)
	} else {
		fail([1,"true", input.Id])
	}
}`),
			out: &Ast{
				Name:  "start",
				Input: "input",
				// TODO
				//Body: Expr{
				//	Assign: &Assign{
				//		Name: "a",
				//		Expr: Expr{
				//			Apply: &Apply{
				//				Name: "ReserveAvailability",
				//				Args: strPtr("input"),
				//			},
				//		},
				//	},
				//},
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

			// For better visual debugging
			//g, _ := Generate(uc.out, &Config{PackageName: "g"})
			//h, _ := Generate(result, &Config{PackageName: "g"})
			//assert.Equal(t, string(g), string(h))
		})
	}
}
