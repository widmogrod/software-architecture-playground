package tictacstatemachine

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMachine(t *testing.T) {
	useCases := map[string]struct {
		commands []Command
		states   []State
		err      []error
	}{
		"game with known possibilities": {
			commands: []Command{
				&MoveCMD{PlayerID: "1", Position: "1.1"}, // game not in progress to make a move
				&CreateGameCMD{FirstPlayerID: "1"},
				&CreateGameCMD{FirstPlayerID: "2"}, // shouldn't be allowed to start game twice
				&JoinGameCMD{SecondPlayerID: "2"},
				&JoinGameCMD{SecondPlayerID: "3"}, // shouldn't be allowed to join game when is full
				&MoveCMD{PlayerID: "1", Position: "1.1"},
				&MoveCMD{PlayerID: "1", Position: "1.1"}, // player makes move twice
				&MoveCMD{PlayerID: "2", Position: "1.1"}, // other player select same position
				&MoveCMD{PlayerID: "2", Position: "2.2"},
				&MoveCMD{PlayerID: "1", Position: "1.2"},
				&MoveCMD{PlayerID: "2", Position: "2.3"},
				&MoveCMD{PlayerID: "1", Position: "1.3"},
				&MoveCMD{PlayerID: "2", Position: "3.3"}, // move after game ended
			},
			err: []error{
				ErrGameNotInProgress,
				nil,
				ErrGameAlreadyStarted,
				nil,
				ErrGameHasAllPlayers,
				nil,
				ErrNotYourTurn,
				ErrPositionTaken,
				nil,
				nil,
				nil,
				nil,
				ErrGameFinished,
			},
			states: []State{
				nil,
				&GameWaitingForPlayer{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID: "1",
					},
				},
				&GameWaitingForPlayer{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID: "1",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
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
					MovesTaken: map[Move]PlayerID{},
					MovesOrder: []Move{},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
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
					MovesTaken: map[Move]PlayerID{},
					MovesOrder: []Move{},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"1.2": {},
						"1.3": {},
						"2.1": {},
						"2.2": {},
						"2.3": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
					},
					MovesOrder: []Move{"1.1"},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"1.2": {},
						"1.3": {},
						"2.1": {},
						"2.2": {},
						"2.3": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
					},
					MovesOrder: []Move{"1.1"},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"1.2": {},
						"1.3": {},
						"2.1": {},
						"2.2": {},
						"2.3": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
					},
					MovesOrder: []Move{"1.1"},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
					AvailableMoves: map[Move]struct{}{
						"1.2": {},
						"1.3": {},
						"2.1": {},
						"2.3": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
					},
					MovesOrder: []Move{"1.1", "2.2"},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"1.3": {},
						"2.1": {},
						"2.3": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
					},
					MovesOrder: []Move{"1.1", "2.2", "1.2"},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
					AvailableMoves: map[Move]struct{}{
						"1.3": {},
						"2.1": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"2.3": "2",
					},
					MovesOrder: []Move{"1.1", "2.2", "1.2", "2.3"},
				},
				&GameResult{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					Winner:         "1",
					WiningSequence: []Move{"1.1", "2.2", "1.2", "2.3", "1.3"},
				},
				&GameResult{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					Winner:         "1",
					WiningSequence: []Move{"1.1", "2.2", "1.2", "2.3", "1.3"},
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			m := NewMachine()
			for i, cmd := range uc.commands {
				m.Handle(cmd)
				assert.Equal(t, uc.states[i], m.State(), "state at index: %d", i)
				assert.Equal(t, uc.err[i], m.LastErr(), "error at index: %d", i)
			}
		})
	}
}
