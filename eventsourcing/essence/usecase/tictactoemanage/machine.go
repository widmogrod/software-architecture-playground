package tictactoemanage

import (
	"errors"
	"fmt"
	"github.com/widmogrod/mkunion/x/machine"
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
	ErrNotExpectedListsOfCommands       = errors.New("not expected lists of commands")
	ErrGameState                        = errors.New("game state error")
)

func Transition(cmd Command, state State) (State, error) {
	return MustMatchCommandR2(cmd,
		func(x *CreateSessionCMD) (State, error) {
			if state != nil {
				return nil, ErrSessionAlreadyCreated
			}

			return &SessionWaitingForPlayers{
				SessionID:    x.SessionID,
				NeedsPlayers: x.NeedsPlayers,
				Players:      []PlayerID{},
			}, nil
		},
		func(x *JoinGameSessionCMD) (State, error) {
			state, ok := state.(*SessionWaitingForPlayers)
			if !ok {
				return nil, ErrSessionNotWaitingForPlayers
			}

			if state.SessionID != x.SessionID {
				return nil, ErrNotTheSameSessions
			}

			for _, player := range state.Players {
				if player == x.PlayerID {
					return nil, ErrPlayerAlreadyJoined
				}
			}

			newState := &SessionWaitingForPlayers{
				SessionID:    state.SessionID,
				NeedsPlayers: state.NeedsPlayers - 1,
				Players:      append(state.Players, x.PlayerID),
			}

			if newState.NeedsPlayers > 0 {
				return newState, nil
			}

			return &SessionReady{
				SessionID: state.SessionID,
				Players:   newState.Players,
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
				if state.SessionID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				var players []PlayerID
				for _, player := range state.Players {
					if player != x.PlayerID {
						players = append(players, player)
					}
				}

				return &SessionWaitingForPlayers{
					SessionID:    state.SessionID,
					NeedsPlayers: state.NeedsPlayers + len(state.Players) - len(players),
					Players:      players,
				}, nil

			case *SessionReady:
				if state.SessionID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				var players []PlayerID
				for _, player := range state.Players {
					if player != x.PlayerID {
						players = append(players, player)
					}
				}

				return &SessionWaitingForPlayers{
					SessionID:    state.SessionID,
					NeedsPlayers: len(players),
					Players:      players,
				}, nil
			}

			panic("not implemented, TODO: error!")
		},
		func(x *NewGameCMD) (State, error) {
			switch state := state.(type) {
			case *SessionReady:
				if state.SessionID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}

				return &SessionInGame{
					SessionID: state.SessionID,
					Players:   state.Players,
					GameID:    x.GameID,
				}, nil

			case *SessionInGame:
				if state.SessionID != x.SessionID {
					return nil, ErrNotTheSameSessions
				}
				//if state.GameID != x.GameID {
				//	panic(ErrNotTheSameGame)
				//}

				return &SessionInGame{
					SessionID: state.SessionID,
					Players:   state.Players,
					GameID:    x.GameID,
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

			if state.SessionID != x.SessionID {
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

			err := game.Handle(action)
			if err != nil {
				return nil, fmt.Errorf("%w, %s", ErrGameState, err)
			}

			if err == nil {
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
								err = game.Handle(&tictacstatemachine.MoveCMD{
									PlayerID: progress.NextMovePlayerID,
									Position: nextMove,
								})
							}
						}
					}
				}
			}

			newState := &SessionInGame{
				SessionID: state.SessionID,
				Players:   state.Players,
				GameID:    state.GameID,
				GameState: game.State(),
			}

			if err != nil {
				msg := err.Error()
				newState.GameProblem = &msg
			}

			return newState, nil
		}, func(x *SequenceCMD) (State, error) {
			if len(x.Commands) < 2 || len(x.Commands) > 5 {
				return nil, ErrNotExpectedListsOfCommands
			}

			var newState = state
			var err error

			for _, cmd := range x.Commands {
				newState, err = Transition(cmd, newState)
				if err != nil {
					return nil, err
				}
			}

			return newState, nil
		})
}

func NewMachine() *machine.Machine[Command, State] {
	return machine.NewSimpleMachine(Transition)
}

func NewMachineWithState(state State) *machine.Machine[Command, State] {
	return machine.NewSimpleMachineWithState(Transition, state)
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
