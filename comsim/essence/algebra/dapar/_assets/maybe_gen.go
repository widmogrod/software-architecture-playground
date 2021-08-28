// GENERATED do not edit!
package _assets

type Maybe interface {
	_unionMaybe()
}

type Just struct {
	T1 Maybe
}

func (_ Just) _unionMaybe() {}

type Nothing struct {}

func (_ Nothing) _unionMaybe() {}
