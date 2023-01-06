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

func UnwrapStateOneOf(x State) State {
	y, ok := x.(*StateOneOf)
	if !ok {
		panic("UnwrapStateOneOf: unexpected value")
	}

	return MustMatchState(y.Unwrap(), func(x *GameWaitingForPlayer) State {
		return x
	}, func(x *GameProgress) State {
		return x
	}, func(x *GameEndWithWin) State {
		return x
	}, func(x *GameEndWithDraw) State {
		return x
	})

}

func WrapCommandOneOf(x Command) *CommandOneOf {
	if x == nil {
		return nil
	}
	return MustMatchCommand(x, func(x *CreateGameCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *JoinGameCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *StartGameCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *MoveCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	})
}

func UnwrapCommandOneOf(x Command) Command {
	y, ok := x.(*CommandOneOf)
	if !ok {
		panic("UnwrapCommandOneOf: unexpected value")
	}

	return MustMatchCommand(y.Unwrap(), func(x *CreateGameCMD) Command {
		return x
	}, func(x *JoinGameCMD) Command {
		return x
	}, func(x *StartGameCMD) Command {
		return x
	}, func(x *MoveCMD) Command {
		return x
	})
}
