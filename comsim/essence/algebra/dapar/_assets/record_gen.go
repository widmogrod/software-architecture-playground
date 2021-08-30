// GENERATED do not edit!
package _assets

type (
	Den interface {
		_unionDen()
	}
	R []RecordR // non-leaf
	RecordR struct {
		Li []In
		R RecordRLiR
	}
	RecordRLiR struct {
		Tu TupleRLiRTu
	}
	TupleRLiRTu struct {
		T1 A
		T2 []B
		T3 RecordRLiRTu
	}
	RecordRLiRTu struct {
		K C
	}
)
func (_ R) _unionDen() {}
