package tictactoeaggregate

import (
	"errors"
	"fmt"
)

func (o *TicTacToeAggregate) Handle(cmd interface{}) error {
	switch c := cmd.(type) {

	case *CreateGameCMD:
		// validate necessary condition
		if o.state != nil {
			return errors.New("Game already exists")
		}

		return o.changes.
			Append(&GameCreated{
				FirstPlayerID: c.FirstPlayerID,
			}).Ok.
			ReduceRecent(o).Err

	case *JoinGameCMD:
		// validate necessary condition
		if o.state == nil {
			return errors.New("Game don't exists")
		}

		if o.state.OneOf.GameWaitingForPlayer == nil {
			return errors.New(fmt.Sprintf("Game must be waiting for player to join game: %#v", c))
		}

		if o.state.OneOf.GameWaitingForPlayer.NeedsPlayers == 0 {
			return errors.New("All player join the game")
		}

		if _, found := o.state.Players[c.SecondPlayerID]; found {
			return errors.New("Player ID already taken")
		}

		return o.changes.
			Append(&SecondPlayerJoined{
				SecondPlayerID: c.SecondPlayerID,
			}).Ok.
			ReduceRecent(o).Err

	case *StartGameCMD:
		// validate necessary condition
		if o.state != nil {
			return errors.New("Game already exists")
		}

		return o.changes.
			Append(&GameCreated{
				FirstPlayerID: c.FirstPlayerID,
			}).Ok.
			Append(&SecondPlayerJoined{
				SecondPlayerID: c.SecondPlayerID,
			}).Ok.
			ReduceRecent(o).Err

	case *MoveCMD:
		// validate necessary condition
		if o.state == nil {
			return errors.New("Game don't exists")
		}

		if o.state.OneOf.GameProgress == nil {
			return errors.New(fmt.Sprintf("Game must be in progress to make a move: %#v", c))
		}

		if o.state.OneOf.GameProgress.NextMovePlayerID != c.PlayerID {
			return errors.New(fmt.Sprintf("Wrong player to make a move: %#v", c))
		}

		if _, ok := o.state.OneOf.GameProgress.AvailableMoves[c.Position]; !ok {
			return errors.New(fmt.Sprintf("Move is not available %#v", c))
		}

		if positions, ok := moveWin(o.state.MovesOrder, c.Position, c.PlayerID); ok {
			return o.changes.
				Append(&Moved{
					PlayerID: c.PlayerID,
					Position: c.Position,
				}).Ok.
				Append(&GameFinish{
					WinnerPlayerID: c.PlayerID,
					Positions:      positions,
				}).Ok.
				ReduceRecent(o).Err
		}

		return o.changes.
			Append(&Moved{
				PlayerID: c.PlayerID,
				Position: c.Position,
			}).Ok.
			ReduceRecent(o).Err
	}

	return errors.New(fmt.Sprintf("Invalid command: %T", cmd))
}

func moveWin(moves []Move, nextMove Move, playerID PlayerID) ([]Move, bool) {
	winseq := [][]Move{
		{"1.1", "1.2", "1.3"},
		{"1.1", "1.3", "1.2"},
		{"1.2", "1.1", "1.3"},
		{"1.2", "1.3", "1.1"},
		{"1.3", "1.1", "1.2"},
		{"1.3", "1.2", "1.1"},
	}

	moves1 := append([]Move{}, moves...)
	moves1 = append(moves1, nextMove)

	a_moves, b_moves := make([]Move, 0), make([]Move, 0)
	idx := 0
	for _, m := range moves1 {
		if idx%2 == 0 {
			a_moves = append(a_moves, m)
		} else {
			b_moves = append(b_moves, m)
		}

		idx++
	}

	for _, seq := range winseq {
		if wonpos, won := is_win_seq(a_moves, seq); won {
			return wonpos, true
		}
		if wonpos, won := is_win_seq(b_moves, seq); won {
			return wonpos, true
		}
	}

	return nil, false
}

func is_win_seq(moves, seq []Move) ([]Move, bool) {
	if len(moves) < len(seq) {
		return nil, false
	}

	if len(moves) == len(seq) {
		for i := 0; i < len(seq); i++ {
			if moves[i] != seq[i] {
				return nil, false
			}
		}

		return seq, true
	}

	offset := len(moves) - 3
	for i := 0; i <= offset; i++ {
		fmt.Println(moves[i:i+3], seq)
		if _, win := is_win_seq(moves[i:i+3], seq); win {
			return seq, true
		}
	}

	return nil, false
}
