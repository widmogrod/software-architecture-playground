package datalang

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestGenerateGo(t *testing.T) {
	// use generated code, as expected and update source code
	// should be set as "true" only during development code generation
	updateSource := false
	useCases := map[string]struct {
		in  string
		out string
	}{
		"maybe_and_list": {
			in:  "_examples/maybe_and_list.dpr",
			out: "_examples/maybe_and_list.go",
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			data, err := ioutil.ReadFile(uc.in)
			if assert.NoError(t, err) {
				ast := MustParse(data)
				gen := MustGenerate(GenerateGo(ast, PackageName("_examples")))
				exp, err := ioutil.ReadFile(uc.out)
				if assert.NoError(t, err) {
					if !assert.Equal(t, string(exp), string(gen)) && updateSource {
						ioutil.WriteFile(uc.out, gen, 0644)
					}
				}
			}
		})
	}
}
