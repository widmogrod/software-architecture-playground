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

		if positions, ok := CheckIfMoveWin(o.state.MovesOrder, c.Position, c.PlayerID); ok {
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
