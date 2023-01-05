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
				ErrPlayerAlreadyJoined,
				ErrNotTheSameSessions,
				nil,
				ErrSessionNotWaitingForPlayers,
				ErrNotTheSameSessions,
				nil,
				ErrNotTheSameSessions,
				ErrNotTheSameGame,
				nil,
			},
			states: []State{
				&SessionWaitingForPlayers{
					ID:           "a",
					NeedsPlayers: 2,
					Players:      []PlayerID{},
				},
				&SessionWaitingForPlayers{
					ID:           "a",
					NeedsPlayers: 2,
					Players:      []PlayerID{},
				},
				&SessionWaitingForPlayers{
					ID:           "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionWaitingForPlayers{
					ID:           "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionWaitingForPlayers{
					ID:           "a",
					NeedsPlayers: 1,
					Players:      []PlayerID{"1"},
				},
				&SessionReady{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
				},
				&SessionReady{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
				},
				&SessionReady{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
				},
				&SessionInGame{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
					GameID:  "g1",
				},
				&SessionInGame{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
					GameID:  "g1",
				},
				&SessionInGame{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
					GameID:  "g1",
				},
				&SessionInGame{
					ID:      "a",
					Players: []PlayerID{"1", "2"},
					GameID:  "g1",
					GameState: &tictacstatemachine.GameProgress{
						TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
							FirstPlayerID:  "1",
							SecondPlayerID: "2",
						},
						NextMovePlayerID: "1",
						AvailableMoves:   tictacstatemachine.NewAvailableMoves(),
						MovesTaken:       map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
						MovesOrder:       []tictacstatemachine.Move{},
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
				assert.Equal(t, uc.states[i], m.State(), "state at index: %d", i)
				assert.Equal(t, uc.err[i], m.LastErr(), "error at index: %d", i)
			}
		})
	}
}
