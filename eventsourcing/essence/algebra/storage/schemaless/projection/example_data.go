package schemaless

type Game struct {
	Players []string
	Winner  string
	IsDraw  bool
}

type SessionsStats struct {
	Wins  int
	Draws int
	Loose int
}
