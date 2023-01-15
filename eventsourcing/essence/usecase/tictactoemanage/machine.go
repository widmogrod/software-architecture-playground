package tictactoemanage

import (
	"errors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/machine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
	"math/rand"
)

const BotPlayerID = "i-am-bot"

var (
	ErrSessionAlreadyCreated            = errors.New("session already created")
	ErrSessionNotWaitingForPlayers      = errors.New("session is not waiting for players to join")
	ErrPlayerAlreadyJoined              = errors.New("player already joined")
	ErrSessionNotReadyToStartGame       = errors.New("session is not ready to start game")
	ErrNotTheSameSessions               = errors.New("not the same sessions")
	ErrSessionNotReadyToAcceptGameInput = errors.New("session is not ready to accept game input")
	ErrNotTheSameGame                   = errors.New("not the same game")
)

func Transition(cmd Command, state State) (State, error) {
	return MustMatchCommandR2(cmd,
		func(x *CreateSessionCMD) (State, error) {
			if state != nil {
				return nil, ErrSessionAlreadyCreated
			}

			return &SessionWaitingForPlayers{
				ID:           x.SessionID,
				NeedsPlayers: x.NeedsPlayers,
				Players:      []PlayerID{},
			}, nil
		},
		func(x *JoinGameSessionCMD) (State, error) {
			state, ok := state.(*SessionWaitingForPlayers)
			if !ok {
				return nil, ErrSessionNotWaitingForPlayers
			}

			if state.ID != x.SessionID {
				return nil, ErrNotTheSameSessions
			}

			for _, player := range state.Players {
				if player == x.PlayerID {
					return nil, ErrPlayerAlreadyJoined
				}
			}

			newState := &SessionWaitingForPlayers{
				ID:           state.ID,
				NeedsPlayers: state.NeedsPlayers - 1,
				Players:      append(state.Players, x.PlayerID),
			}

			if newState.NeedsPlayers > 0 {
				return newState, nil
			}

			return &SessionReady{
				ID:      state.ID,
				Players: newState.Players,
			}, nil
		},
		func(x *GameSessionWithBotCMD) (State, error) {
			return Transition(&JoinGameSessionCMD{
				SessionID: x.SessionID,
				PlayerID:  BotPlayerID,
			}, state)
		},
		func(x *LeaveGameSessionCMD) (State, error) {
			switch state := state.(type) {
			case *SessionWaitingForPlayers:
				if state.ID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				var players []PlayerID
				for _, player := range state.Players {
					if player != x.PlayerID {
						players = append(players, player)
					}
				}

				return &SessionWaitingForPlayers{
					ID:           state.ID,
					NeedsPlayers: state.NeedsPlayers + float64(len(state.Players)-len(players)),
					Players:      players,
				}, nil

			case *SessionReady:
				if state.ID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				var players []PlayerID
				for _, player := range state.Players {
					if player != x.PlayerID {
						players = append(players, player)
					}
				}

				return &SessionWaitingForPlayers{
					ID:           state.ID,
					NeedsPlayers: float64(len(players)),
					Players:      players,
				}, nil
			}

			panic("not implemented, TODO: error!")
		},
		func(x *NewGameCMD) (State, error) {
			switch state := state.(type) {
			case *SessionReady:
				if state.ID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				return &SessionInGame{
					ID:      state.ID,
					Players: state.Players,
					GameID:  x.GameID,
				}, nil

			case *SessionInGame:
				if state.ID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}
				//if state.GameID != x.GameID {
				//	panic(ErrNotTheSameGame)
				//}

				return &SessionInGame{
					ID:      state.ID,
					Players: state.Players,
					GameID:  x.GameID,
				}, nil

			default:
				return nil, ErrSessionNotReadyToStartGame
			}
		},
		func(x *GameActionCMD) (State, error) {
			state, ok := state.(*SessionInGame)
			if !ok {
				return nil, ErrSessionNotReadyToAcceptGameInput
			}

			if state.ID != x.SessionID {
				return nil, ErrNotTheSameSessions
			}

			if state.GameID != x.GameID {
				return nil, ErrNotTheSameGame
			}

			game := tictacstatemachine.NewMachineWithState(state.GameState)
			//if game.State() == nil {
			//	game.Handle(&tictacstatemachine.StartGameCMD{
			//		FirstPlayerID:  state.Players[0],
			//		SecondPlayerID: state.Players[1],
			//	})
			//}
			action := x.Action

			if sg, ok := action.(*tictacstatemachine.StartGameCMD); ok {
				if HasBotPlayer(state.Players) ||
					rand.Float64() < 0.5 {
					sg.FirstPlayerID = state.Players[0]
					sg.SecondPlayerID = state.Players[1]
				} else {
					sg.FirstPlayerID = state.Players[1]
					sg.SecondPlayerID = state.Players[0]
				}
			}

			game.Handle(action)

			if game.LastErr() == nil {
				if _, ok := action.(*tictacstatemachine.MoveCMD); ok {
					if progress, ok := game.State().(*tictacstatemachine.GameProgress); ok {
						if IsBot(progress.NextMovePlayerID) {
							//nextMove := tictactoeaggregate.SlidingNextMoveMinMax(
							nextMove := tictactoeaggregate.NextMoveMinMax(
								progress.MovesOrder,
								progress.BoardRows,
								progress.BoardCols,
							)

							if nextMove != "" {
								game.Handle(&tictacstatemachine.MoveCMD{
									PlayerID: progress.NextMovePlayerID,
									Position: nextMove,
								})
							}
						}
					}
				}
			}

			//if tictacstatemachine.IsGameFinished(game.State()) {
			//	return &SessionReady{
			//		ID:      state.ID,
			//		Players: state.Players,
			//	}
			//}

			newState := &SessionInGame{
				ID:        state.ID,
				Players:   state.Players,
				GameID:    state.GameID,
				GameState: game.State(),
			}

			if game.LastErr() != nil {
				msg := game.LastErr().Error()
				newState.GameProblem = &msg
			}

			return newState, nil
		})
}

func NewMachine() *machine.Machine[Command, State] {
	return machine.NewSimpleMachine(Transition)
}

func IsBot(id PlayerID) bool {
	return id == BotPlayerID
}

func HasBotPlayer(players []PlayerID) bool {
	for _, player := range players {
		if IsBot(player) {
			return true
		}
	}

	return false
}
