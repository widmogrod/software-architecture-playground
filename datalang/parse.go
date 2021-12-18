package datalang

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	def = lexer.MustSimple([]lexer.Rule{
		//{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`, nil},
		{"IdentLower", `[a-z_][a-zA-Z0-9_]*`, nil},
		{"IdentUpper", `[A-Z_][a-zA-Z0-9_]*`, nil},
		{"EOL", `[\n\r]+`, nil},
		//{"whitespace", `\s+`, nil},
		{"whitespace", `[ \t]+`, nil},
		{"comment", `(?i)rem[^\n]*\n`, nil},
		//{"String", `"(\\"|[^"])*"`, nil},
		//{"Number", `[-+]?(\d*\.)?\d+`, nil},
	})
	parser = participle.MustBuild(&Ast{}) //participle.Lexer(def),
	//participle.Elide("comment"),

)

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
