package dapar

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func Parse(in []byte) (*Ast, error) {
	p := &Parser{
		ast: &Ast{},
	}
	err := p.Parse(in)
	return p.ast, err
}

func MustParse(in []byte) *Ast {
	ast, err := Parse(in)
	if err != nil {
		panic(err)
	}
	return ast
}

type Ast struct {
	DataTypes []DataType
}

type DataType struct {
	Name string
	Sum  []DataConstructor
}

type DataConstructor struct {
	Name string
	Args []Typ
}

type Typ struct {
	Name string
}

type Parser struct {
	ast *Ast
}

func (p *Parser) Parse(in []byte) error {
	tok1, rest1 := p.Select(in)
	switch t1 := tok1.(type) {
	case *Ident:
		tok2, rest2 := p.Select(rest1)
		switch tok2.(type) {
		case *Eq:
			dc, err := p.ParseDataConstructors(rest2)
			if err != nil {
				panic(err)
			}
			p.ast.DataTypes = append(p.ast.DataTypes, DataType{
				Name: t1.String(),
				Sum:  dc,
			})
		case nil:
			// noop
			return fmt.Errorf("cannot make sence with: %#v; what left: %s", tok1, rest1)
		}
	case nil:
		// noop
	default:
		return fmt.Errorf("cannot make sence with first token: %#v in input %s", tok1, in)
	}

	return nil
}

func (p *Parser) ParseDataConstructors(in []byte) ([]DataConstructor, error) {
	var res []DataConstructor

	tok1, rest1 := p.Select(in)
	switch t1 := tok1.(type) {
	case *Or:
		dcs, err := p.ParseDataConstructors(rest1)
		if err != nil {
			panic(err)
		}
		res = append(res, dcs...)

	case *Ident:
		tok2, rest2 := p.Select(rest1)
		switch tok2.(type) {
		case *Or:
			dc := DataConstructor{
				Name: t1.String(),
			}
			res = append(res, dc)

			dcs, err := p.ParseDataConstructors(rest2)
			if err != nil {
				panic(err)
			}
			res = append(res, dcs...)

		case *POpen:
			types, rest3, err := p.ParseArgs(rest2)
			if err != nil {
				panic(err)
			}

			dc := DataConstructor{
				Name: t1.String(),
				Args: types,
			}
			res = append(res, dc)

			dcs, err := p.ParseDataConstructors(rest3)
			if err != nil {
				panic(string(rest3))
				panic(err)
			}
			res = append(res, dcs...)

		case nil:
			dc := DataConstructor{
				Name: t1.String(),
			}
			res = append(res, dc)
			break

		default:
			return nil, fmt.Errorf("expects data constuctor or nothing more, but get %#v: %s", tok2, in)
		}
	default:
		return nil, fmt.Errorf("no data constructor: %s", in)
	}

	return res, nil
}

func (p *Parser) ParseArgs(in []byte) ([]Typ, []byte, error) {
	var res []Typ
	var left []byte

	tok1, rest1 := p.Select(in)
	switch t1 := tok1.(type) {
	case *Ident:
		t := Typ{
			Name: t1.String(),
		}
		res = append(res, t)

		types, rest, err := p.ParseArgs(rest1)
		if err != nil {
			panic(err)
		}
		res = append(res, types...)
		left = rest

	case *Comma:
		types, rest, err := p.ParseArgs(rest1)
		if err != nil {
			panic(err)
		}
		res = append(res, types...)
		left = rest

	case *PClose, nil:
		// noop
		left = rest1
		break

	default:
		return nil, left, fmt.Errorf("expects type or nothing more, but get %#v: %s", tok1, in)
	}
	return res, left, nil
}

type Token = interface {
	Token()
}

var (
	ident      = regexp.MustCompile(`^\w+`)
	whitespace = regexp.MustCompile(`^\s+`)
)

type (
	Eq     struct{}
	Or     struct{}
	POpen  struct{}
	PClose struct{}
	Comma  struct{}

	Ident struct {
		found []byte
	}
)

func (_ Eq) Token()     {}
func (_ Or) Token()     {}
func (_ POpen) Token()  {}
func (_ PClose) Token() {}
func (_ Comma) Token()  {}
func (_ Ident) Token()  {}

func (i Ident) String() string {
	return string(i.found)
}

func (p *Parser) Select(in []byte) (Token, []byte) {
	if l := len(in); l == 0 {
		return nil, nil
	}

	res := whitespace.Find(in)
	if l := len(res); l > 0 {
		return p.Select(in[l:])
	}

	res = ident.Find(in)
	if l := len(res); l > 0 {
		return &Ident{found: res}, in[l:]
	}

	if in[0] == 0x3D { // =
		return &Eq{}, in[1:]
	}
	if in[0] == 0x7C { // |
		return &Or{}, in[1:]
	}
	if in[0] == 0x28 { // (
		return &POpen{}, in[1:]
	}
	if in[0] == 0x29 { // )
		return &PClose{}, in[1:]
	}
	if in[0] == 0x2C { // ,
		return &Comma{}, in[1:]
	}

	return nil, in
}

type Config struct {
	PackageName string
}

func Generate(ast *Ast, c *Config) ([]byte, error) {
	result := &bytes.Buffer{}

	fmt.Fprintf(result, "// GENERATED do not edit!\n")
	fmt.Fprintf(result, "package %s\n\n", c.PackageName)

	for _, dt := range ast.DataTypes {
		fmt.Fprintf(result, "type %s interface {\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "	%sDataType()\n", strings.Title(dt.Name))
		fmt.Fprintf(result, "}\n")
		for _, dc := range dt.Sum {
			fmt.Fprintf(result, "\ntype %s struct {", strings.Title(dc.Name))
			for _, t := range dc.Args {
				fmt.Fprintf(result, "\n\t%s %s\n", strings.Title(t.Name), strings.Title(t.Name))
			}
			fmt.Fprintf(result, "}\n")
			fmt.Fprintf(result, "\nfunc (_ %s) %sDataType() {}\n", strings.Title(dc.Name), strings.Title(dt.Name))
		}
	}

	return result.Bytes(), nil
}
