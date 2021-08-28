// GENERATED do not edit!
package _assets

type Err interface {
	_unionErr()
}

type Ok struct {}

func (_ Ok) _unionErr() {}

type Err struct {}

func (_ Err) _unionErr() {}

type Faults interface {
	_unionFaults()
}

type IOFault struct {}

func (_ IOFault) _unionFaults() {}

func (_ IOFault) _unionErr() {} // Alias

type Unexpected struct {}

func (_ Unexpected) _unionFaults() {}

func (_ Unexpected) _unionErr() {} // Alias
