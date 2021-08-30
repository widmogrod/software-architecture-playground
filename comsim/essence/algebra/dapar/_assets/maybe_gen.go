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
