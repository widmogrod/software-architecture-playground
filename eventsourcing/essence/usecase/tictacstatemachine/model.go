package tictacstatemachine

// Value objects, values that restrict cardinality of the state.
type (
	PlayerID = string
	Move     = string
)

// Commands that trigger state transitions.
//
//go:generate mkunion -name=Command
type (
	CreateGameCMD struct{ FirstPlayerID PlayerID }
	JoinGameCMD   struct{ SecondPlayerID PlayerID }
	StartGameCMD  struct {
		FirstPlayerID  PlayerID
		SecondPlayerID PlayerID
	}
	MoveCMD struct {
		PlayerID PlayerID
		Position Move
	}
)

// State of the game.
// Commands are used to update or change state
//
//go:generate mkunion -name=State
type (
	GameWaitingForPlayer struct {
		TicTacToeBaseState
	}

	GameProgress struct {
		TicTacToeBaseState

		NextMovePlayerID Move
		AvailableMoves   map[Move]struct{}
		MovesTaken       map[Move]PlayerID
		MovesOrder       []Move
	}

	GameEndWithWin struct {
		TicTacToeBaseState

		Winner         PlayerID
		WiningSequence []Move
		MovesTaken     map[Move]PlayerID
	}
	GameEndWithDraw struct {
		TicTacToeBaseState

		MovesTaken map[Move]PlayerID
	}
)

type TicTacToeBaseState struct {
	FirstPlayerID  PlayerID
	SecondPlayerID PlayerID
}
