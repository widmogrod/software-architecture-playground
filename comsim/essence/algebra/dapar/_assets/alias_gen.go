// GENERATED do not edit!
package _assets

type (
	Err interface {
		_unionErr()
	}
	Ok struct {}
)
func (_ Ok) _unionErr() {}

type ErrVisitor interface {
	VisitOk(x Ok) interface{}
	VisitErr(x Err) interface{}
}

func MapErr(value Err, v ErrVisitor) interface{} {
	switch x := value.(type) {
	case Ok:
		return v.VisitOk(x)
	case Err:
		return v.VisitErr(x)
	default:
		panic(`unknown type`)
	}
}

type (
	Faults interface {
		_unionFaults()
	}
	IOFault struct {}
	Unexpected struct {}
)
func (_ IOFault) _unionFaults() {}
func (_ IOFault) _unionErr() {} // Alias
func (_ Unexpected) _unionFaults() {}
func (_ Unexpected) _unionErr() {} // Alias

type FaultsVisitor interface {
	VisitIOFault(x IOFault) interface{}
	VisitUnexpected(x Unexpected) interface{}
}

func MapFaults(value Faults, v FaultsVisitor) interface{} {
	switch x := value.(type) {
	case IOFault:
		return v.VisitIOFault(x)
	case Unexpected:
		return v.VisitUnexpected(x)
	default:
		panic(`unknown type`)
	}
}
