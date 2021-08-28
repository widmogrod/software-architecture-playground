package dapar

import (
	"fmt"
	"regexp"
)

func NaiveParse(in []byte) (*Ast, error) {
	p := &NaiveParser{
		ast: &Ast{},
	}
	err := p.Parse(in)
	return p.ast, err
}

type NaiveParser struct {
	ast *Ast
}

func (p *NaiveParser) Parse(in []byte) error {
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

func (p *NaiveParser) ParseDataConstructors(in []byte) ([]DataConstructor, error) {
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
			_, rest3, err := p.ParseArgs(rest2)
			if err != nil {
				panic(err)
			}

			dc := DataConstructor{
				Name: t1.String(),
				//Args: types,
			}
			res = append(res, dc)

			dcs, err := p.ParseDataConstructors(rest3)
			if err != nil {
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

func (p *NaiveParser) ParseArgs(in []byte) ([]Typ, []byte, error) {
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

func (p *NaiveParser) Select(in []byte) (Token, []byte) {
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
