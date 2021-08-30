// GENERATED do not edit!
package _assets

type (
	Nestrecord interface {
		_unionNestrecord()
	}
	R struct {
		A RARecord
	}
	RARecord struct {
		B RABRecord
	}
	RABRecord struct {
		C RABCRecord
	}
	RABCRecord struct {
		A E
	}
)
func (_ R) _unionNestrecord() {}
