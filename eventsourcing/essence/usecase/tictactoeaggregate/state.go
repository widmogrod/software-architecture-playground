package tictactoeaggregate

type PlayerID = string
type Move = string

type (
	GameWaitingForPlayer struct {
		NeedsPlayers int32
	}

	GameProgress struct {
		NextMovePlayerID Move
		AvailableMoves   map[Move]struct{}
	}

	GameResult struct {
		Winner         PlayerID
		WiningSequence []Move
	}

	TicTacToeState struct {
		Players    map[PlayerID]struct{}
		MovesTaken map[Move]PlayerID
		MovesOrder []Move
		OneOf      struct {
			GameWaitingForPlayer *GameWaitingForPlayer
			GameProgress         *GameProgress
			GameResult           *GameResult
		}
	}
)
