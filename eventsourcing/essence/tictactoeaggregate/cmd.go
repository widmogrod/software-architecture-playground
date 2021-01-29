package tictactoeaggregate

type CreateGameCMD struct {
	FirstPlayerID string
}

type GameCreated struct {
	FirstPlayerID string
}

type JoinGameCMD struct {
	SecondPlayerID string
}

type SecondPlayerJoined struct {
	SecondPlayerID string
}

type StartGameCMD struct {
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
