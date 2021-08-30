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

type DataVisitor interface {
	VisitMany(x Many) interface{}
	VisitMore(x More) interface{}
}

func MapData(value Data, v DataVisitor) interface{} {
	switch x := value.(type) {
	case Many:
		return v.VisitMany(x)
	case More:
		return v.VisitMore(x)
	default:
		panic(`unknown type`)
	}
}
