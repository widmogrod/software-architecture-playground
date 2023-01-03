package tictacstatemachine

import (
	"errors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
)

var (
	ErrGameAlreadyStarted = errors.New("game already started")
	ErrGameHasAllPlayers  = errors.New("game is not waiting for player")
	ErrGameNotInProgress  = errors.New("cannot move, game is not in progress")
	ErrNotYourTurn        = errors.New("not your turn")
	ErrPositionTaken      = errors.New("position is taken")
	ErrGameFinished       = errors.New("game is finished")
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

	if _, ok := o.state.(*GameResult); ok {
		// Game is over, no more commands can be applied
		panic(ErrGameFinished)
	}

	o.state = MustMatchCommand(
		cmd,
		func(x *CreateGameCMD) State {
			if o.state != nil {
				panic(ErrGameAlreadyStarted)
			}

			return &GameWaitingForPlayer{
				TicTacToeBaseState: TicTacToeBaseState{
					FirstPlayerID: x.FirstPlayerID,
				},
			}
		},
		func(x *JoinGameCMD) State {
			state, ok := o.state.(*GameWaitingForPlayer)
			if !ok {
				panic(ErrGameHasAllPlayers)
			}

			base := state.TicTacToeBaseState
			base.SecondPlayerID = x.SecondPlayerID

			return &GameProgress{
				TicTacToeBaseState: base,
				MovesTaken:         map[Move]PlayerID{},
				MovesOrder:         []Move{},
				NextMovePlayerID:   state.FirstPlayerID,
				AvailableMoves: map[Move]struct{}{
					"1.1": {},
					"1.2": {},
					"1.3": {},
					"2.1": {},
					"2.2": {},
					"2.3": {},
					"3.1": {},
					"3.2": {},
					"3.3": {},
				},
			}
		},
		func(x *StartGameCMD) State {
			o.Handle(&CreateGameCMD{FirstPlayerID: x.FirstPlayerID})
			o.Handle(&JoinGameCMD{SecondPlayerID: x.SecondPlayerID})
			return o.state
		},
		func(x *MoveCMD) State {
			state, ok := o.state.(*GameProgress)
			if !ok {
				panic(ErrGameNotInProgress)
			}

			if state.NextMovePlayerID != x.PlayerID {
				panic(ErrNotYourTurn)
			}

			if _, ok := state.AvailableMoves[x.Position]; !ok {
				panic(ErrPositionTaken)
			}

			delete(state.AvailableMoves, x.Position)
			state.MovesTaken[x.Position] = x.PlayerID
			state.MovesOrder = append(state.MovesOrder, x.Position)

			if x.PlayerID == state.FirstPlayerID {
				state.NextMovePlayerID = state.SecondPlayerID
			} else {
				state.NextMovePlayerID = state.FirstPlayerID
			}

			// Check if there is a winner
			if _, win := tictactoeaggregate.CheckIfMoveWin(state.MovesOrder, x.Position, x.PlayerID); win {
				return &GameResult{
					TicTacToeBaseState: state.TicTacToeBaseState,
					Winner:             x.PlayerID,
					WiningSequence:     state.MovesOrder,
				}
			} else if len(state.AvailableMoves) == 0 {
				return &GameResult{
					TicTacToeBaseState: state.TicTacToeBaseState,
				}
			}

			return state
		},
	)
}
