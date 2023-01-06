package tictactoemanage

import (
	"errors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
)

var (
	ErrSessionAlreadyCreated            = errors.New("session already created")
	ErrSessionNotWaitingForPlayers      = errors.New("session is not waiting for players to join")
	ErrPlayerAlreadyJoined              = errors.New("player already joined")
	ErrSessionNotReadyToStartGame       = errors.New("session is not ready to start game")
	ErrNotTheSameSessions               = errors.New("not the same sessions")
	ErrSessionNotReadyToAcceptGameInput = errors.New("session is not ready to accept game input")
	ErrNotTheSameGame                   = errors.New("not the same game")
)

func NewMachine() *Machine {
	return &Machine{
		state: nil,
	}
}

type Machine struct {
	state   State
	lastErr error
}

func (o *Machine) State() State {
	return o.state
}

func (o *Machine) LastErr() error {
	return o.lastErr
}

func (o *Machine) Handle(cmd Command) {
	o.lastErr = nil

	defer func() {
		if r := recover(); r != nil {
			o.lastErr = r.(error)
		}
	}()

	o.state = MustMatchCommand(cmd,
		func(x *CreateSessionCMD) State {
			if o.state != nil {
				panic(ErrSessionAlreadyCreated)
			}

			return &SessionWaitingForPlayers{
				ID:           x.SessionID,
				NeedsPlayers: x.NeedsPlayers,
				Players:      []PlayerID{},
			}
		}, func(x *JoinGameSessionCMD) State {
			state, ok := o.state.(*SessionWaitingForPlayers)
			if !ok {
				panic(ErrSessionNotWaitingForPlayers)
			}

			if state.ID != x.SessionID {
				panic(ErrNotTheSameSessions)
			}

			for _, player := range state.Players {
				if player == x.PlayerID {
					panic(ErrPlayerAlreadyJoined)
				}
			}

			newState := &SessionWaitingForPlayers{
				ID:           state.ID,
				NeedsPlayers: state.NeedsPlayers - 1,
				Players:      append(state.Players, x.PlayerID),
			}

			if newState.NeedsPlayers > 0 {
				return newState
			}

			return &SessionReady{
				ID:      state.ID,
				Players: newState.Players,
			}
		},
		func(x *LeaveGameSessionCMD) State {
			switch state := o.state.(type) {
			case *SessionWaitingForPlayers:
				if state.ID != x.SessionID {
					panic(ErrNotTheSameSessions)
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
				}

			case *SessionReady:
				if state.ID != x.SessionID {
					panic(ErrNotTheSameSessions)
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
				}
			}

			panic("not implemented, TODO: error!")
		},
		func(x *NewGameCMD) State {
			state, ok := o.state.(*SessionReady)
			if !ok {
				panic(ErrSessionNotReadyToStartGame)
			}
			if state.ID != x.SessionID {
				panic(ErrNotTheSameSessions)
			}

			return &SessionInGame{
				ID:      state.ID,
				Players: state.Players,
				GameID:  x.GameID,
			}

		}, func(x *GameActionCMD) State {
			state, ok := o.state.(*SessionInGame)
			if !ok {
				panic(ErrSessionNotReadyToAcceptGameInput)
			}

			if state.ID != x.SessionID {
				panic(ErrNotTheSameSessions)
			}

			if state.GameID != x.GameID {
				panic(ErrNotTheSameGame)
			}

			game := tictacstatemachine.NewMachineWithState(state.GameState)
			//if game.State() == nil {
			//	game.Handle(&tictacstatemachine.StartGameCMD{
			//		FirstPlayerID:  state.Players[0],
			//		SecondPlayerID: state.Players[1],
			//	})
			//}
			if action, ok := x.Action.(*tictacstatemachine.CommandOneOf); ok {
				game.Handle(action.Unwrap())
			} else {
				game.Handle(x.Action)
			}

			if tictacstatemachine.IsGameFinished(game.State()) {
				return &SessionReady{
					ID:      state.ID,
					Players: state.Players,
				}
			}

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

			return newState
		})
}
