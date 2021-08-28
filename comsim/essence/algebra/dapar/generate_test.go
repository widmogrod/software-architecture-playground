package dapar

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

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
		"should generate nothing on input = 'maybe = just(a) | nothing`'": {
			c:    c,
			ast:  MustParse([]byte(`maybe = just(maybe) | nothing`)),
			file: "_assets/maybe_gen.go",
		},
		"should generate nothing on input = 'data = many([in]) | more [to]`'": {
			c:    c,
			ast:  MustParse([]byte(`data = many([in]) | more [to]`)),
			file: "_assets/complex_gen.go",
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

func TestGenTypes(t *testing.T) {
	useCases := map[string]struct {
		typ      *Typ
		expected []byte
	}{
		"a": {
			typ:      ref(tuple(typ("a"))),
			expected: []byte("struct {\n\tT1 A\n}\n"),
		},
		"b": {
			typ:      ref(tuple(typ("a"), list("b"))),
			expected: []byte("struct {\n\tT1 A\n\tT2 []B\n}\n"),
		},
		"c": {
			typ: &Typ{List: &Typ{Tuple: []*Typ{
				ref(list("a")),
				ref(typ("b")),
			}}},
			expected: []byte("[]struct {\n\tT1 []A\n\tT2 B\n}\n"),
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			res := bytes.NewBuffer(nil)
			gentypes(res, uc.typ, 0)
			assert.Equal(t, string(uc.expected), string(res.Bytes()))
		})
	}
}
