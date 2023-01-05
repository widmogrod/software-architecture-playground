package tictactoemanage

import "github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"

type (
	SessionID = string
	GameID    = string
	PlayerID  = string
)

//go:generate mkunion -name=Command
type (
	CreateSessionCMD struct {
		SessionID    SessionID
		NeedsPlayers int
	}
	JoinGameSessionCMD struct {
		SessionID SessionID
		PlayerID  PlayerID
	}
	NewGameCMD struct {
		SessionID SessionID
		GameID    GameID
	}
	GameActionCMD struct {
		SessionID SessionID
		GameID    GameID
		Action    tictacstatemachine.Command
	}
)

//go:generate mkunion -name=State
type (
	SessionWaitingForPlayers struct {
		ID           SessionID
		NeedsPlayers int
		Players      []PlayerID
	}
	SessionReady struct {
		ID      SessionID
		Players []PlayerID
	}
	SessionInGame struct {
		ID            SessionID
		Players       []PlayerID
		GameID        GameID
		GameState     tictacstatemachine.State
		GameProblem   *string
		PreviousGames []GameID
	}
)
