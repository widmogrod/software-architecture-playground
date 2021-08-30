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
