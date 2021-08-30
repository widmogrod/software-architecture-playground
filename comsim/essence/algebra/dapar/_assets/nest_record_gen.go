// GENERATED do not edit!
package _assets

type (
	Nestrecord interface {
		_unionNestrecord()
	}
	R struct {
		A RecordRA
	}
	RecordRA struct {
		B RecordRAB
	}
	RecordRAB struct {
		C RecordRABC
	}
	RecordRABC struct {
		A E
	}
)
func (_ R) _unionNestrecord() {}
