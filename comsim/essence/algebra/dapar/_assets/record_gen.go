// GENERATED do not edit!
package _assets

type (
	Den interface {
		_unionDen()
	}
	R []RRecord // non-leaf
	RRecord struct {
		Li []In
		R RLiRRecord
	}
	RLiRRecord struct {
		Tu RLiRTuTuple
	}
	RLiRTuTuple struct {
		T1 A
		T2 []B
		T3 RLiRTuRecord
	}
	RLiRTuRecord struct {
		K C
	}
)
func (_ R) _unionDen() {}

type DenVisitor interface {
	VisitR(x R) interface{}
}

func MapDen(value Den, v DenVisitor) interface{} {
	switch x := value.(type) {
	case R:
		return v.VisitR(x)
	default:
		panic(`unknown type`)
	}
}
