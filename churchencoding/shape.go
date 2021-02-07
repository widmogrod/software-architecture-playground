package churchencoding

import (
	"math"
)

type shape = float64
type (
	Circle    = func(float64, float64, float64) shape
	Rectangle = func(float64, float64, float64, float64) shape
	Shape     = func(Circle, Rectangle) shape
)

func _Circle(x, y, r float64) Shape {
	return func(circle Circle, _ Rectangle) shape {
		return circle(x, y, r)
	}
}

func _Rectangle(x, y, w, h float64) Shape {
	return func(_ Circle, rectangle Rectangle) shape {
		return rectangle(x, y, w, h)
	}
}

func area(s Shape) float64 {
	return s(func(x, y, r float64) shape {
		return math.Pi * r * r
	}, func(x, y, w, h float64) shape {
		return w * h
	})
}
