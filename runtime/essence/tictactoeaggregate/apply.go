package tictactoeaggregate

import (
	"errors"
	"fmt"
)

func (o *TicTacToeAggregate) Apply(change interface{}) error {
	switch c := change.(type) {
	case *GameCreated:
		if o.state != nil {
			return errors.New("order cannot be created game twice, check your logic")
		}

		// when everything is ok, record changes that you want to make
		o.state = &TicTacToeState{
			Players:    make(map[PlayerID]struct{}),
			MovesTaken: make(map[Move]PlayerID),
			MovesOrder: make([]Move, 0),
		}

		o.state.Players[c.FirstPlayerID] = struct{}{}
		o.state.OneOf = struct {
			GameWaitingForPlayer *GameWaitingForPlayer
			GameProgress         *GameProgress
			GameResult           *GameResult
		}{
			GameWaitingForPlayer: &GameWaitingForPlayer{NeedsPlayers: 1},
		}

	case *SecondPlayerJoined:
		if o.state == nil {
			return errors.New("order cannot be created game twice, check your logic")
		}

		o.state.Players[c.SecondPlayerID] = struct{}{}
		o.state.OneOf.GameWaitingForPlayer = nil
		o.state.OneOf.GameResult = nil
		o.state.OneOf.GameProgress = &GameProgress{
			NextMovePlayerID: getNext(o.state.Players, c.SecondPlayerID),
			AvailableMoves: map[Move]struct{}{
				"1.1": struct{}{},
				"1.2": struct{}{},
				"1.3": struct{}{},
				"2.1": struct{}{},
				"2.2": struct{}{},
				"2.3": struct{}{},
				"3.1": struct{}{},
				"3.2": struct{}{},
				"3.3": struct{}{},
			},
		}

	case *Moved:
		if o.state == nil {
			return errors.New("Cannot make a move when game don't start")
		}

		if c.PlayerID == o.state.OneOf.GameProgress.NextMovePlayerID {
			o.state.MovesTaken[c.Position] = c.PlayerID
			o.state.MovesOrder = append(o.state.MovesOrder, c.Position)

			delete(o.state.OneOf.GameProgress.AvailableMoves, c.Position)

			o.state.OneOf.GameProgress.NextMovePlayerID = getNext(o.state.Players, c.PlayerID)
		} else {
			return errors.New("Wrong player move, from replyied events")
		}

	case *GameFinish:
		if o.state == nil {
			return errors.New("Cannot finish game that don't started")
		}

		o.state.OneOf.GameWaitingForPlayer = nil
		o.state.OneOf.GameProgress = nil
		o.state.OneOf.GameResult = &GameResult{
			Winner:         c.WinnerPlayerID,
			WiningSequence: c.Positions,
		}

	default:
		return errors.New(fmt.Sprintf("unsupported type to handle %T", change))
	}

	return nil
}

func getNext(pm map[PlayerID]struct{}, id PlayerID) (res PlayerID) {
	for playerID := range pm {
		if playerID == id {
			continue
		} else {
			res = playerID
		}
	}

	return
}
