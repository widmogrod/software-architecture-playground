package datalang

type (
	Ast struct {
		Datas []Data `("data" @@)+`
	}
	Data struct {
		Name string        `@Ident`
		Poli []string      `("(" @Ident | ("," @Ident)* ")")*`
		Body []Constructor `(("=" @@) | ("|" @@))+`
	}
	Constructor struct {
		Name  string   `@Ident`
		Value []string `("(" @Ident ("," @Ident)* ")")?`
	}
)

type (
	Value struct {
		Name  string  ` [a-z]`
		Tuple []Value `| "(" @@ ("," @@)* ")"`
		//Map   []Map   `| "{" @@ ("," @@)* "}"`
		//List  []Value `| "[" @@ ("," @@)* "]"`
	}

	//Map struct {
	//	Key  string `    @IdentUpper`
	//	Type string `":" @IdentUpper`
	//}
)
