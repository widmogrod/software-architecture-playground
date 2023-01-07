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
