// GENERATED do not edit!
package _assets

type Data interface {
	_unionData()
}

type Many struct {
	T1 []In
}

func (_ Many) _unionData() {}

type More []To

func (_ More) _unionData() {}
