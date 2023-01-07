package tictacstatemachine

func WrapStateOneOf(x State) *StateOneOf {
	if x == nil {
		return nil
	}

	return MustMatchState(x, func(x *GameWaitingForPlayer) *StateOneOf {
		return MapStateToOneOf(x)
	}, func(x *GameProgress) *StateOneOf {
		return MapStateToOneOf(x)
	}, func(x *GameEndWithWin) *StateOneOf {
		return MapStateToOneOf(x)
	}, func(x *GameEndWithDraw) *StateOneOf {
		return MapStateToOneOf(x)
	})
}
