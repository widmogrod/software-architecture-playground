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

	return rule.Eval(data)
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

func (p *Predicate) Eval(data interface{}) error {
	if p.Type != nil {
		switch data.(type) {
		case int, int8, int16, int32, int64:
			if *p.Type != TypeInt {
				// wrap error to provide more info
				return fmt.Errorf("%w: %v", ErrWrongType, *p.Type)
			}

		case bool:
			if *p.Type != TypeBool {
				// wrap error to provide more info
				return fmt.Errorf("%w: %v", ErrWrongType, *p.Type)
			}

		case string:
			if *p.Type != TypeString {
				// wrap error to provide more info
				return fmt.Errorf("%w: %v", ErrWrongType, *p.Type)
			}
		}

		// if si map, check if the field is in the map
		if *p.Type == TypeMap {
			// check if the field is in the map
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			if v.Kind() != reflect.Map {
				// wrap error to provide more info
				return fmt.Errorf("%w: %v", ErrWrongType, *p.Type)
			}
		}

		return nil
	}

	if p.Eq != nil {
		if reflect.DeepEqual(p.Eq, data) {
			return nil
		}
		// wrap error to provide more info
		return fmt.Errorf("%w: %v != %v", ErrValueNotEqual, p.Eq, data)
	}

	if p.In != nil {
		for _, v := range p.In {
			if reflect.DeepEqual(v, data) {
				return nil
			}
		}
		// wrap error to provide more info
		return fmt.Errorf("%w: %v", ErrValueNotContainedIn, data)
	}

	if p.Fields != nil {
		mapAny, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%w: %v", ErrValueNotMap, reflect.TypeOf(data).String())
		}

		for k, v := range p.Fields {
			if _, ok := mapAny[k]; !ok {
				return fmt.Errorf("%w: %v", ErrFieldInMap, k)
			}
			err := v.Eval(mapAny[k])
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

var typToString = map[Typ]string{
	TypeInt:    "int",
	TypeString: "string",
	TypeBool:   "bool",
}

func (p *Predicate) String() string {
	if p.Type != nil {
		return fmt.Sprintf("Typ(%s)", typToString[*p.Type])
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
