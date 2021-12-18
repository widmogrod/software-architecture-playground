package datalang

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type UseCases = map[string]Spec

type Spec struct {
	in  []byte
	out *Ast
	err error
}

func TestParser(t *testing.T) {
	useCases := UseCases{
		"should return AST with a few data types": {
			in: []byte(`

data maybe(a) = Nothing | Some(a)

data list
	| Cons(a, list)
	| Nil

`),
			out: &Ast{
				Datas: []Data{
					{
						Name: "maybe",
						Poli: []string{"a"},
						Body: []Constructor{
							{Name: "Nothing"},
							{Name: "Some", Value: []string{"a"}},
						},
					},
					{
						Name: "list",
						Body: []Constructor{
							{Name: "Cons", Value: []string{"a", "list"}},
							{Name: "Nil"},
						},
					},
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

			t.Logf("Formated:\n%s", Fmt(uc.out))
			t.Logf("Generated:\n%s", MustGenerate(GenerateGo(uc.out)))
			assert.Equal(t, uc.out, result)
		})
	}
}
