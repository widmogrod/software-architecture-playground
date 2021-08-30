// GENERATED do not edit!
package _assets

type (
	Err interface {
		_unionErr()
	}
	Ok struct {}
)
func (_ Ok) _unionErr() {}

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
