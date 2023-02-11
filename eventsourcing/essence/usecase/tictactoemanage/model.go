package tictactoemanage

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
)

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

		// TODO: PlayerID should be set on server!
		PlayerID PlayerID
	}
	GameSessionWithBotCMD struct {
		SessionID SessionID
	}
	LeaveGameSessionCMD struct {
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
	// SequenceCMD helps to address problem where
	// individual commands acn be sent in order 1,2,
	// but they are processed in order 2,1
	SequenceCMD struct {
		Commands []Command
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
		ID          SessionID
		Players     []PlayerID
		GameID      GameID
		GameState   tictacstatemachine.State
		GameProblem *string
	}
)

//go:generate mkunion -name=Query
type (
	SessionStatsQuery struct {
		SessionID SessionID
	}
)

//go:generate mkunion -name=QueryResult
type (
	SessionStatsResult struct {
		ID         SessionID
		TotalGames int
		TotalDraws int
		PlayerWins map[PlayerID]int
	}
)
