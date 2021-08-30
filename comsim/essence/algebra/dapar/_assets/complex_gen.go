// GENERATED do not edit!
package _assets

type (
	Data interface {
		_unionData()
	}
	Many struct {
		T1 []In
	}
	More []To // leaf
)
func (_ Many) _unionData() {}
func (_ More) _unionData() {}
