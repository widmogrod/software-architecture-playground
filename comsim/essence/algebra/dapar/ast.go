package dapar

type (
	Ast struct {
		DataTypes []DataType `@@*`
	}
	DataType struct {
		Name string            `@Ident "="`
		Sum  []DataConstructor `(@@ ("|" @@)* ";"?)`
	}
	DataConstructor struct {
		Name string `@Ident`
		Args *Typ   `@@*`
	}
	Typ struct {
		Name   string   `  @Ident`
		List   *Typ     `| "[" @@ "]"`
		Tuple  []*Typ   `| "(" @@ ("," @@)* ")"`
		Record []Record `| "{" @@ ("," @@)* "}"`
	}
	Record struct {
		Key   string `@Ident ":"`
		Value *Typ   `@@`
	}
)
