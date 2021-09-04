package wokpar

type (
	Ast struct {
		Name  string `"flow" @Ident`
		Input string `"(" @Ident ")"`
		Body  []Expr `"{" @@+ "}"`
	}
	Expr struct {
		End    *End    `   @@`
		Apply  *Apply  ` | @@`
		Assign *Assign ` | @@`
		Choose *Choose ` | @@`
	}
	Apply struct {
		Name string    `( @Ident`
		Args *Selector `  "(" @@* ")")`
	}
	Assign struct {
		Name VarName `@@ "="`
		Expr Expr    `@@`
	}
	Choose struct {
		Predicate Predicate ` "if" @@`
		Then      []Expr    ` "{" @@+ "}"`
		Else      []Expr    `("else" "{" @@+ "}")*`
	}
	VarName struct {
		Ignore bool    `  @"_"`
		Name   *string `| @Ident`
	}
	End struct {
		Result *Return ` "return" @@?`
		Fail   *Fail   `|  "fail" @@?`
	}
)

type (
	Return struct {
		Args *Selector `"(" @@* ")"`
	}
	Fail struct {
		Args *Selector `"(" @@* ")"`
	}
)

type (
	Eq struct {
		Left  Selector `    @@+`
		Right Selector `"," @@+`
	}
	And struct {
		Left  Predicate `    @@`
		Right Predicate `"," @@`
	}
	Or struct {
		Left  Predicate `    @@+`
		Right Predicate `"," @@+`
	}
	Predicate struct {
		Eq     *Eq       `       "eq" "(" @@ ")"`
		And    *And      `|     "and" "(" @@ ")"`
		Or     *And      `|      "or" "(" @@ ")"`
		Exists *Selector `|  "exists" "(" @@ ")"`
	}
	Selector struct {
		SetValue *Value   `  @@`
		GetValue []string `| (@Ident ("." @Ident)*)+`
	}
)

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type (
	Value struct {
		Float  *float64   `  @Float`
		Int    *int       `| @Int`
		String *string    `| @String`
		Bool   *Boolean   `| @("true" | "false")`
		Map    []Map      `| "{" @@ ("," @@)* "}"`
		List   []Selector `| "[" @@ ("," @@)* "]"`
	}
	Map struct {
		Key   Selector `    @@`
		Value Selector `":" @@`
	}
)
