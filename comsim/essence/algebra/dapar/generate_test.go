package dapar

import (
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
		"should generate = 'maybe = just(a) | nothing`'": {
			c:    c,
			ast:  MustParse([]byte(`maybe = just(maybe) | nothing`)),
			file: "_assets/maybe_gen.go",
		},
		"should generate = 'data = many([in]) | more [to]`'": {
			c:    c,
			ast:  MustParse([]byte(`data = many([in]) | more [to]`)),
			file: "_assets/complex_gen.go",
		},
		"should generate = 'den = r {list:[in], r: {tu:(a,[b],{k:c})}}`'": {
			c:    c,
			ast:  MustParse([]byte(`den = r [{li:[in], r: {tu:(a,[b],{k:c})}}]`)),
			file: "_assets/record_gen.go",
		},
		"should generate = 'nestrecord = r {a: {b: {c: {a: e}}}}`'": {
			c:    c,
			ast:  MustParse([]byte(`nestrecord = r {a: {b: {c: {a: e}}}}`)),
			file: "_assets/nest_record_gen.go",
		},
		"should generate = 'err = Ok | Err = faults; faults...`'": {
			c: c,
			ast: MustParse([]byte(`
err = Ok | Err = faults;
faults = IOFault | Unexpected
`)),
			file: "_assets/alias_gen.go",
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
