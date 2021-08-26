package dapar

import (
	"fmt"
	"regexp"
)

func Parse(in []byte) (*Ast, error) {
	p := &Parser{
		ast: &Ast{},
	}
	err := p.Parse(in)
	return p.ast, err
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
			return fmt.Errorf("cannot make sence with: %#v, %s", tok1, rest1)
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
			types, err := p.ParseArgs(rest2)
			if err != nil {
				panic(err)
			}

			dc := DataConstructor{
				Name: t1.String(),
				Args: types,
			}
			res = append(res, dc)

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

func (p *Parser) ParseArgs(in []byte) ([]Typ, error) {
	var res []Typ

	tok1, rest1 := p.Select(in)
	switch t1 := tok1.(type) {
	case *Ident:
		t := Typ{
			Name: t1.String(),
		}
		res = append(res, t)

		types, err := p.ParseArgs(rest1)
		if err != nil {
			panic(err)
		}
		res = append(res, types...)

	case *Comma:
		types, err := p.ParseArgs(rest1)
		if err != nil {
			panic(err)
		}
		res = append(res, types...)

	case *PClose, nil:
		// noop
		break

	default:
		return nil, fmt.Errorf("expects type or nothing more, but get %#v: %s", tok1, in)
	}
	return res, nil
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
