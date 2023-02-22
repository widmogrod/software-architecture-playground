package tictactoemanage

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"testing"
)

func TestNewMachine(t *testing.T) {
	useCases := map[string]struct {
		commands []Command
		states   []State
		err      []error
	}{
		"all know session scenarios": {
			commands: []Command{
				&CreateSessionCMD{
					SessionID:    "a",
					NeedsPlayers: 2,
				},
				&CreateSessionCMD{ // cannot start session two times
					SessionID:    "a",
					NeedsPlayers: 2,
				},
				&LeaveGameSessionCMD{
					SessionID: "a",
					PlayerID:  "1",
				},
				&JoinGameSessionCMD{
					SessionID: "a",
					PlayerID:  "1",
				},
				&JoinGameSessionCMD{ // cannot join session with same player id
					SessionID: "a",
					PlayerID:  "1",
				},
				&JoinGameSessionCMD{ // cannot join different session
					SessionID: "b",
					PlayerID:  "2",
				},
				&JoinGameSessionCMD{
					SessionID: "a",
					PlayerID:  "2",
				},
				&JoinGameSessionCMD{ // cannot join session that is not waiting for players
					SessionID: "a",
					PlayerID:  "3",
				},
				&LeaveGameSessionCMD{
					SessionID: "a",
					PlayerID:  "1",
				},
				&JoinGameSessionCMD{
					SessionID: "a",
					PlayerID:  "1",
				},
				&NewGameCMD{ // cannot start game in different session
					SessionID: "b",
					GameID:    "g1",
				},
				&NewGameCMD{
					SessionID: "a",
					GameID:    "g1",
				},
				&GameActionCMD{ // cannot make action in different session
					SessionID: "bbb",
					GameID:    "g1",
					Action: &tictacstatemachine.StartGameCMD{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
				},
				&GameActionCMD{ // session has different game that action issued
					SessionID: "a",
					GameID:    "ggggggg",
					Action: &tictacstatemachine.StartGameCMD{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
				},
				&GameActionCMD{
					SessionID: "a",
					GameID:    "g1",
					Action: &tictacstatemachine.StartGameCMD{
						FirstPlayerID:  "1",
						SecondPlayerID: "2",
					},
				},
			},
			err: []error{
				nil,
				ErrSessionAlreadyCreated,
				nil,
				nil,
				ErrPlayerAlreadyJoined,
				ErrNotTheSameSessions,
				nil,
				ErrSessionNotWaitingForPlayers,
				nil,
				nil,
				ErrNotTheSameSessions,
				nil,
				ErrNotTheSameSessions,
				ErrNotTheSameGame,
				nil,
			},
			states: []State{
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 2,
					Players:      []PlayerID{},
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 2,
					Players:      []PlayerID{},
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 2,
					Players:      []PlayerID(nil),
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionReady{
					SessionID: "a",
					Players:   []PlayerID{"1", "2"},
				},
				&SessionReady{
					SessionID: "a",
					Players:   []PlayerID{"1", "2"},
				},
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"2"},
				},
				&SessionReady{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
				},
				&SessionReady{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
				},
				&SessionInGame{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
					GameID:    "g1",
				},
				&SessionInGame{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
					GameID:    "g1",
				},
				&SessionInGame{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
					GameID:    "g1",
				},
				&SessionInGame{
					SessionID: "a",
					Players:   []PlayerID{"2", "1"},
					GameID:    "g1",
					GameState: &tictacstatemachine.GameProgress{
						TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
							FirstPlayerID:  "1",
							SecondPlayerID: "2",
							BoardRows:      3,
							BoardCols:      3,
							WinningLength:  3,
						},
						NextMovePlayerID: "1",
						MovesTaken:       map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
						MovesOrder:       []tictacstatemachine.Move{},
					},
				},
			},
		},

		"sequence of commands must have at least two": {
			commands: []Command{
				&SequenceCMD{
					Commands: []Command{
						&CreateSessionCMD{
							SessionID:    "s1",
							NeedsPlayers: 2,
						},
					},
				},
			},
			states: []State{
				nil,
			},
			err: []error{
				ErrNotExpectedListsOfCommands,
			},
		},
		"sequence of commands works": {
			commands: []Command{
				&SequenceCMD{
					Commands: []Command{
						&CreateSessionCMD{
							SessionID:    "a",
							NeedsPlayers: 2,
						},
						&JoinGameSessionCMD{
							SessionID: "a",
							PlayerID:  "1",
						},
					},
				},
			},
			states: []State{
				&SessionWaitingForPlayers{
					SessionID:    "a",
					NeedsPlayers: 1,
					Players: []PlayerID{
						"1",
					},
				},
			},
			err: []error{
				nil,
			},
		},
	}

	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			m := NewMachine()
			for i, cmd := range uc.commands {
				err := m.Handle(cmd)
				assert.Equal(t, uc.states[i], m.State(), "state at index: %d", i)
				assert.Equal(t, uc.err[i], err, "error at index: %d", i)
			}
		})
	}
}
