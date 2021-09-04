package wokpar

import (
	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild(&Ast{})

func Parse(in []byte) (*Ast, error) {
	res := &Ast{}
	return res, parser.ParseBytes("", in, res)
}

func MustParse(in []byte) *Ast {
	ast, err := Parse(in)
	if err != nil {
		panic(err)
	}
	return ast
}
