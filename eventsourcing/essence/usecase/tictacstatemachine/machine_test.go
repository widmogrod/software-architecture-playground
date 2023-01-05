package tictacstatemachine

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
	"testing"
)

func TestNewMachine(t *testing.T) {
	useCases := map[string]struct {
		commands []Command
		states   []State
		err      []error
	}{
		"cannot join game with same player id": {
			commands: []Command{
				&CreateGameCMD{
					FirstPlayerID: "1",
				},
				&JoinGameCMD{
					SecondPlayerID: "1",
				},
			},
			err: []error{
				nil,
				ErrUniquePlayers,
			},
			states: []State{
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
			},
		},
		"game with known possibilities, and winning sequence": {
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
				&GameEndWithWin{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					Winner:         "1",
					WiningSequence: []Move{"1.1", "1.2", "1.3"},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"2.3": "2",
						"1.3": "1",
					},
				},
				&GameEndWithWin{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					Winner:         "1",
					WiningSequence: []Move{"1.1", "1.2", "1.3"},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"2.3": "2",
						"1.3": "1",
					},
				},
			},
		},
		"game with known possibilities, and tie sequence": {
			commands: []Command{
				&StartGameCMD{
					FirstPlayerID:  "1",
					SecondPlayerID: "2",
				},
				&MoveCMD{PlayerID: "1", Position: "1.1"},
				&MoveCMD{PlayerID: "2", Position: "2.2"},
				&MoveCMD{PlayerID: "1", Position: "1.2"},
				&MoveCMD{PlayerID: "2", Position: "1.3"},
				&MoveCMD{PlayerID: "1", Position: "2.3"},
				&MoveCMD{PlayerID: "2", Position: "2.1"},
				&MoveCMD{PlayerID: "1", Position: "3.1"},
				&MoveCMD{PlayerID: "2", Position: "3.2"},
				&MoveCMD{PlayerID: "1", Position: "3.3"},
			},
			err: []error{
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			},
			states: []State{
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
					MovesOrder: []Move{
						"1.1",
					},
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
					MovesOrder: []Move{
						"1.1", "2.2",
					},
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
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
					AvailableMoves: map[Move]struct{}{
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
						"1.3": "2",
					},
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
						"1.3",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"2.1": {},
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"1.3": "2",
						"2.3": "1",
					},
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
						"1.3", "2.3",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
					AvailableMoves: map[Move]struct{}{
						"3.1": {},
						"3.2": {},
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"1.3": "2",
						"2.3": "1",
						"2.1": "2",
					},
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
						"1.3", "2.3", "2.1",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "2",
					AvailableMoves: map[Move]struct{}{
						"3.3": {},
						"3.2": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"1.3": "2",
						"2.3": "1",
						"2.1": "2",
						"3.1": "1",
					},
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
						"1.3", "2.3", "2.1",
						"3.1",
					},
				},
				&GameProgress{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					NextMovePlayerID: "1",
					AvailableMoves: map[Move]struct{}{
						"3.3": {},
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"1.3": "2",
						"2.3": "1",
						"2.1": "2",
						"3.1": "1",
						"3.2": "2",
					},
					MovesOrder: []Move{
						"1.1", "2.2", "1.2",
						"1.3", "2.3", "2.1",
						"3.1", "3.2",
					},
				},
				&GameEndWithDraw{
					TicTacToeBaseState: TicTacToeBaseState{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
					MovesTaken: map[Move]PlayerID{
						"1.1": "1",
						"2.2": "2",
						"1.2": "1",
						"1.3": "2",
						"2.3": "1",
						"2.1": "2",
						"3.1": "1",
						"3.2": "2",
						"3.3": "1",
					},
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			m := NewMachine()
			for i, cmd := range uc.commands {
				m.Handle(cmd)
				if s, ok := m.State().(*GameProgress); ok {
					tictactoeaggregate.PrintGame(s.MovesTaken)
				}
				assert.Equal(t, uc.states[i], m.State(), "state at index: %d", i)
				assert.Equal(t, uc.err[i], m.LastErr(), "error at index: %d", i)
			}
		})
	}
}
