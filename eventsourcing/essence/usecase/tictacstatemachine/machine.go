package tictacstatemachine

import (
	"errors"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
)

var (
	ErrGameAlreadyStarted = errors.New("game already started")
	ErrGameHasAllPlayers  = errors.New("game is not waiting for player")
	ErrUniquePlayers      = errors.New("game can not have same player twice")
	ErrGameNotInProgress  = errors.New("cannot move, game is not in progress")
	ErrNotYourTurn        = errors.New("not your turn")
	ErrPositionTaken      = errors.New("position is taken")
	ErrGameFinished       = errors.New("game is finished")
	ErrInputInvalid       = errors.New("input is invalid")
)

func NewMachine() *Machine {
	return NewMachineWithState(nil)
}

func NewMachineWithState(s State) *Machine {
	return &Machine{
		state: s,
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

	if _, ok := o.state.(*GameEndWithWin); ok {
		// Game is over, no more commands can be applied
		panic(ErrGameFinished)
	}

	o.state = MustMatchCommand(
		cmd,
		func(x *CreateGameCMD) State {
			if o.state != nil {
				panic(ErrGameAlreadyStarted)
			}

			rows, cols, length := GameRules(x.BoardRows, x.BoardCols, x.WinningLength)

			return &GameWaitingForPlayer{
				TicTacToeBaseState: TicTacToeBaseState{
					FirstPlayerID: x.FirstPlayerID,
					BoardRows:     rows,
					BoardCols:     cols,
					WinningLength: length,
				},
			}
		},
		func(x *JoinGameCMD) State {
			state, ok := o.state.(*GameWaitingForPlayer)
			if !ok {
				panic(ErrGameHasAllPlayers)
			}

			if state.FirstPlayerID == x.SecondPlayerID {
				panic(ErrUniquePlayers)
			}

			base := state.TicTacToeBaseState
			base.SecondPlayerID = x.SecondPlayerID

			return &GameProgress{
				TicTacToeBaseState: base,
				MovesTaken:         map[Move]PlayerID{},
				MovesOrder:         []Move{},
				NextMovePlayerID:   state.FirstPlayerID,
			}
		},
		func(x *StartGameCMD) State {
			o.Handle(&CreateGameCMD{
				FirstPlayerID: x.FirstPlayerID,
				BoardRows:     x.BoardRows,
				BoardCols:     x.BoardCols,
				WinningLength: x.WinningLength,
			})
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

			move, err := ParsePosition(x.Position, state.BoardRows, state.BoardCols)
			if err != nil {
				panic(err)
			}

			if _, ok := state.MovesTaken[move]; ok {
				panic(ErrPositionTaken)
			}

			state.MovesTaken[x.Position] = x.PlayerID
			state.MovesOrder = append(state.MovesOrder, move)

			if x.PlayerID == state.FirstPlayerID {
				state.NextMovePlayerID = state.SecondPlayerID
			} else {
				state.NextMovePlayerID = state.FirstPlayerID
			}

			// Check if there is a winner
			winseq := tictactoeaggregate.GenerateWiningPositions(state.WinningLength, state.BoardRows, state.BoardCols)
			if seq, win := tictactoeaggregate.CheckIfMoveWin(state.MovesOrder, winseq); win {
				return &GameEndWithWin{
					TicTacToeBaseState: state.TicTacToeBaseState,
					Winner:             x.PlayerID,
					WiningSequence:     seq,
					MovesTaken:         state.MovesTaken,
				}
			} else if len(state.MovesTaken) == (state.BoardRows * state.BoardCols) {
				return &GameEndWithDraw{
					TicTacToeBaseState: state.TicTacToeBaseState,
					MovesTaken:         state.MovesTaken,
				}
			}

			return state
		},
	)
}

func ParsePosition(position Move, boardRows int, boardCols int) (Move, error) {
	var r, c int
	_, err := fmt.Sscanf(position, "%d.%d", &r, &c)
	if err != nil {
		return "", fmt.Errorf("move cannot be parsed %w; %s", ErrInputInvalid, err)
	}

	if r < 1 ||
		c < 1 ||
		r > boardRows ||
		c > boardCols {
		return "", fmt.Errorf("move position is out of bounds %w", ErrInputInvalid)
	}

	return tictactoeaggregate.MkMove(r, c), nil

}

func GameRules(rows int, cols int, length int) (int, int, int) {
	r, c, l := rows, cols, length

	max := 10

	if l < 3 {
		l = 3
	} else if l > max {
		l = max
	}

	if r <= l {
		r = l
	}

	if c <= l {
		c = l
	}

	return r, c, l
}

func NewAvailableMoves(rows, cols int) map[Move]struct{} {
	m := map[Move]struct{}{}
	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			m[fmt.Sprintf("%d.%d", i, j)] = struct{}{}
		}
	}
	return m
}

func IsGameFinished(x State) bool {
	switch x.(type) {
	case *GameEndWithDraw,
		*GameEndWithWin:
		return true
	}

	return false
}
