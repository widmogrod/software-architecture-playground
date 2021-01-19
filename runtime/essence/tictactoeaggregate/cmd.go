package tictactoeaggregate

type StartGameCMD struct {
	FirstPlayerID  string
	SecondPlayerID string
}

type GameStarted struct {
	FirstPlayerID  string
	SecondPlayerID string
}

type MoveCMD struct {
	PlayerID string
	Position string
}

type Moved struct {
	PlayerID string
	Position string
}

type GameFinish struct {
	WinnerPlayerID string
	Positions      []string
}
