package churchencoding

import (
	"math"
)

//type T = interface{}
//
//type Visitor interface {
//	VisitCreateGame(tictactoeaggregate.CreateGameCMD) T
//	VisitJoinGame(tictactoeaggregate.JoinGameCMD) T
//	VisitStartGame(tictactoeaggregate.StartGameCMD) T
//	VisitMove(tictactoeaggregate.MoveCMD) T
//}

type shape = float64
type Shape = func(
	func(float64, float64, float64) shape,
	func(float64, float64, float64, float64) shape,
) shape

func _Circle(x, y, r float64) Shape {
	return func(circle func(float64, float64, float64) shape, _ func(float64, float64, float64, float64) shape) shape {
		return circle(x, y, r)
	}
}

func _Rectangle(x, y, w, h float64) Shape {
	return func(_ func(float64, float64, float64) shape, rectangle func(float64, float64, float64, float64) shape) shape {
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
