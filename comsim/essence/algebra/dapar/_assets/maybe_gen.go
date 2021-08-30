// GENERATED do not edit!
package _assets

type (
	Maybe interface {
		_unionMaybe()
	}
	Just struct {
		T1 Maybe
	}
	Nothing struct {}
)
func (_ Just) _unionMaybe() {}
func (_ Nothing) _unionMaybe() {}

type MaybeVisitor interface {
	VisitJust(x Just) interface{}
	VisitNothing(x Nothing) interface{}
}

func MapMaybe(value Maybe, v MaybeVisitor) interface{} {
	switch x := value.(type) {
	case Just:
		return v.VisitJust(x)
	case Nothing:
		return v.VisitNothing(x)
	default:
		panic(`unknown type`)
	}
}
