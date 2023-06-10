package gm

type Type uint8

const (
	StringType Type = iota
	IntType
	BoolType
)

func TypeToString(t Type) string {
	switch t {
	case StringType:
		return "string"
	case IntType:
		return "int64"
	case BoolType:
		return "bool"
	}
	panic("unknown type")
}

type AttrType struct {
	//Guard      *Predicate
	T          Type
	Required   bool
	Identifier bool
	Default    interface{}
}
