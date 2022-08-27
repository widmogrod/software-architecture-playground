package gm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// iota of type
const (
	TypeInt Typ = iota
	TypeString
	TypeBool
	TypeMap
)

func PtrType(x Typ) *Typ {
	return &x
}

var typToString = map[Typ]string{
	TypeInt:    "int",
	TypeString: "string",
	TypeBool:   "bool",
}

func (t Typ) String() string {
	return typToString[t]
}

type (
	RuleID = string

	Typ uint

	Guard struct {
		rules map[string]*Predicate
		lock  sync.RWMutex
	}

	Predicate struct {
		Type   *Typ
		Eq     interface{}
		In     []interface{}
		Fields map[string]Predicate
		And    []Predicate
		Or     []Predicate
	}
)

func NewGuard() *Guard {
	return &Guard{
		rules: make(map[string]*Predicate),
	}
}

var (
	ErrRuleNotFound        = errors.New("rule not found")
	ErrRuleAlreadyIdExists = errors.New("rule id already exists")
)

func (a *Guard) CreateRule(id RuleID, predicate Predicate) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if _, ok := a.rules[id]; ok {
		// wrap error to provide more info
		return fmt.Errorf("%w: %v", ErrRuleAlreadyIdExists, id)
	}

	a.rules[id] = &predicate
	return nil
}

func (a *Guard) GetRule(id RuleID) (*Predicate, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if _, ok := a.rules[id]; ok {
		return a.rules[id], nil
	}

	return nil, fmt.Errorf("%w: %v", ErrRuleNotFound, id)
}

func (a *Guard) MergePredicates(x, y Predicate) Predicate {
	if x.Fields != nil && y.Fields != nil {
		// get and merge common fields
		commonFields := make(map[string]Predicate)
		for k, v := range x.Fields {
			if _, ok := y.Fields[k]; ok {
				commonFields[k] = a.MergePredicates(v, y.Fields[k])
			}
		}
		// get and merge common and different fields
		for k, v := range y.Fields {
			if _, ok := commonFields[k]; !ok {
				commonFields[k] = a.MergePredicates(v, y.Fields[k])
			}
		}
		return Predicate{Fields: commonFields}
	}

	if x.And != nil && y.And != nil {
		return Predicate{And: append(x.And, y.And...)}
	}

	if y.And != nil {
		return Predicate{And: append(x.And, y)}
	}

	return Predicate{And: append(x.And, y)}
}

func (a *Guard) CreteRuleBaseOf(id RuleID, baseId RuleID, predicate Predicate) error {
	baseRule, err := a.GetRule(baseId)
	if err != nil {
		return err
	}
	finalRule := a.MergePredicates(*baseRule, predicate)

	return a.CreateRule(id, finalRule)
}

func (a *Guard) EvalRule(id RuleID, data interface{}) error {
	rule, err := a.GetRule(id)
	if err != nil {
		return err
	}

	return rule.Eval(&GolangTypeReader{data: data})
}

var (
	ErrWrongType                = errors.New("wrong type")
	ErrValueNotContainedIn      = errors.New("value not contained in")
	ErrValueNotEqual            = errors.New("value not equal")
	ErrValueNotMap              = errors.New("value not map, but looking for field")
	ErrFieldInMap               = errors.New("the field is not in the map")
	ErrOneOfAndPredicatesFailed = errors.New("one of the AND predicates failed")
	ErrAllOrPredicatesFailed    = errors.New("all of the OR predicates failed")
)

var _ DataReader = &GolangTypeReader{}

type GolangTypeReader struct {
	data interface{}
}

func (g *GolangTypeReader) GetType() Typ {
	switch g.data.(type) {
	case int, int8, int16, int32, int64:
		return TypeInt
	case bool:
		return TypeBool

	case string:
		return TypeString
	}

	// check if the field is in the map
	v := reflect.ValueOf(g.data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Map {
		return TypeMap
	}

	panic("not implemented")
}

func (g *GolangTypeReader) Equals(eq interface{}) bool {
	return reflect.DeepEqual(g.data, eq)
}

func (g *GolangTypeReader) ToKeyable() (DataReader, error) {
	if _, ok := g.data.(map[string]interface{}); ok {
		return g, nil
	}

	v := reflect.ValueOf(g.data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Map {
		// convert map to map[string]interface{}
		m := make(map[string]interface{})
		for _, k := range v.MapKeys() {
			m[k.String()] = v.MapIndex(k).Interface()
		}
		return &GolangTypeReader{data: m}, nil
	}

	if v.Kind() == reflect.Struct {
		// convert struct to map[string]interface{}
		m := make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			name, ok := v.Type().Field(i).Tag.Lookup("name")
			if !ok {
				name = v.Type().Field(i).Name
			}
			m[name] = v.Field(i).Interface()
		}
		return &GolangTypeReader{data: m}, nil
	}

	return nil, ErrValueNotMap
}

func (g *GolangTypeReader) GetKey(key string) (data DataReader, exists bool) {
	if m, ok := g.data.(map[string]interface{}); ok {
		if v, ok := m[key]; ok {
			return &GolangTypeReader{data: v}, true
		}
	}

	return nil, false
}

type (
	DataReader interface {
		GetType() Typ
		Equals(eq interface{}) bool
		ToKeyable() (DataReader, error)
		GetKey(key string) (data DataReader, exists bool)
	}
)

func (p *Predicate) Eval(data DataReader) error {
	if p.Type != nil {
		if data.GetType() != *p.Type {
			return fmt.Errorf("%w: given %v expeteed %s", ErrWrongType, data.GetType().String(), p.Type.String())
		}
		return nil

	}

	if p.Eq != nil {
		if data.Equals(p.Eq) {
			return nil
		}
		// wrap error to provide more info
		return fmt.Errorf("%w: %v != %s", ErrValueNotEqual, p.Eq, data.GetType())
	}

	if p.In != nil {
		for _, v := range p.In {
			if data.Equals(v) {
				return nil
			}
		}
		// wrap error to provide more info
		return fmt.Errorf("%w: %s", ErrValueNotContainedIn, data.GetType())
	}

	if p.Fields != nil {
		m, err := data.ToKeyable()
		if err != nil {
			return fmt.Errorf("%w: %s", ErrValueNotMap, data.GetType())
		}

		for k, v := range p.Fields {
			item, exists := m.GetKey(k)
			if !exists {
				return fmt.Errorf("%w: %v", ErrFieldInMap, k)
			}
			err := v.Eval(item)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if p.And != nil {
		for _, v := range p.And {
			err := v.Eval(data)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrOneOfAndPredicatesFailed, v.String())
			}
		}
		return nil
	}

	if p.Or != nil {
		for _, v := range p.Or {
			err := v.Eval(data)
			if err == nil {
				return nil
			}
		}
		return fmt.Errorf("%w: %v", ErrAllOrPredicatesFailed, p.String())
	}

	return errors.New(fmt.Sprintf("eval: type %s not implemented", reflect.TypeOf(data).String()))
}

func (p *Predicate) String() string {
	if p.Type != nil {
		return fmt.Sprintf("Typ(%s)", p.Type.String())
	}

	if p.Eq != nil {
		// get type of p.Eq
		return fmt.Sprintf("Eq(%v)", reflect.TypeOf(p.Eq).String())
	}

	if p.In != nil {
		return fmt.Sprintf("In(%v)", reflect.TypeOf(p.In).String())
	}

	if p.Fields != nil {
		// get list of fields
		fields := make([]string, 0, len(p.Fields))
		for k, v := range p.Fields {
			fields = append(fields, fmt.Sprintf("%s=%s", k, v.String()))
		}
		return fmt.Sprintf("Fields(%v)", strings.Join(fields, ","))
	}

	if p.And != nil {
		// get list of and predicates
		predicates := make([]string, 0, len(p.And))
		for _, v := range p.And {
			predicates = append(predicates, v.String())
		}
		return fmt.Sprintf("And(%v)", strings.Join(predicates, ","))
	}

	if p.Or != nil {
		// get list of or predicates
		predicates := make([]string, 0, len(p.Or))
		for _, v := range p.Or {
			predicates = append(predicates, v.String())
		}
		return fmt.Sprintf("Or(%v)", strings.Join(predicates, ","))
	}

	panic("unreachable")
}
