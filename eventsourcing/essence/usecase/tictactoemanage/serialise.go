package tictactoemanage

import "github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"

func WrapStateOneOf(x State) *StateOneOf {
	if x == nil {
		return nil
	}

	return MustMatchState(x, func(x *SessionWaitingForPlayers) *StateOneOf {
		return MapStateToOneOf(x)
	}, func(x *SessionReady) *StateOneOf {
		return MapStateToOneOf(x)
	}, func(x *SessionInGame) *StateOneOf {
		return &StateOneOf{
			SessionInGame: &SessionInGame{
				ID:            x.ID,
				Players:       x.Players,
				GameID:        x.GameID,
				GameState:     tictacstatemachine.WrapStateOneOf(x.GameState),
				GameProblem:   x.GameProblem,
				PreviousGames: x.PreviousGames,
			},
		}
	})
}

func UnwrapStateOneOf(x State) State {
	y, ok := x.(*StateOneOf)
	if !ok {
		panic("UnwrapStateOneOf: unexpected value")
	}
	return MustMatchState(y.Unwrap(), func(x *SessionWaitingForPlayers) State {
		return x
	}, func(x *SessionReady) State {
		return x
	}, func(x *SessionInGame) State {
		y := x
		y.GameState = tictacstatemachine.UnwrapStateOneOf(y.GameState)
		return y
	})
}

func WrapCommandOneOf(x Command) *CommandOneOf {
	if x == nil {
		return nil
	}
	return MustMatchCommand(x, func(x *CreateSessionCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *JoinGameSessionCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *LeaveGameSessionCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *NewGameCMD) *CommandOneOf {
		return MapCommandToOneOf(x)
	}, func(x *GameActionCMD) *CommandOneOf {
		return &CommandOneOf{
			GameActionCMD: &GameActionCMD{
				SessionID: x.SessionID,
				GameID:    x.GameID,
				Action:    tictacstatemachine.WrapCommandOneOf(x.Action),
			},
		}
	})
}

func UnwrapCommandOneOf(x Command) Command {
	y, ok := x.(*CommandOneOf)
	if !ok {
		panic("UnwrapStateOneOf: unexpected value")
	}

	return MustMatchCommand(y.Unwrap(), func(x *CreateSessionCMD) Command {
		return x
	}, func(x *JoinGameSessionCMD) Command {
		return x
	}, func(x *LeaveGameSessionCMD) Command {
		return x
	}, func(x *NewGameCMD) Command {
		return x
	}, func(x *GameActionCMD) Command {
		y := x
		y.Action = tictacstatemachine.UnwrapCommandOneOf(y.Action)
		return y
	})
}
