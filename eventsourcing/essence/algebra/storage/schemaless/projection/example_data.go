package projection

type Game struct {
	SessionID string
	Players   []string
	Winner    string
	IsDraw    bool
}

type SessionsStats struct {
	Wins  int
	Draws int
	Loose int
}
