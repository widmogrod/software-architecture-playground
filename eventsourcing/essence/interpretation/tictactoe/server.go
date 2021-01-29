package tictactoe

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe/proto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ proto.TicTacToeAggregateServer = &tictactoeServer{}

type storer interface {
	NewAggregate(ctx context.Context, aggregateID string, handle aggregate.HandleFunc) (aggregate.Aggregate, error)
	MutateAggregate(ctx context.Context, aggregateID string, handle aggregate.HandleFunc) (aggregate.Aggregate, error)
}

func NewTicTacToeServer(store storer) *tictactoeServer {
	return &tictactoeServer{
		store: store,
	}
}

type tictactoeServer struct {
	store storer
}

func (s *tictactoeServer) CreateGame(ctx context.Context, request *proto.CreateGameRequest) (*proto.CreateGameResponse, error) {
	agg, err := s.store.NewAggregate(ctx, request.GameID, func(agg aggregate.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.CreateGameCMD{
			FirstPlayerID: request.FirstPlayerID,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error CreateGame(%s)! details %s", request.GameID, err.Error(),
		)
	}

	return &proto.CreateGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) JoinGame(ctx context.Context, request *proto.JoinGameRequest) (*proto.JoinGameResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggregate.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.JoinGameCMD{
			SecondPlayerID: request.SecondPlayerID,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error JoinGame(%s)! details %s", request.GameID, err.Error(),
		)
	}

	return &proto.JoinGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) Move(ctx context.Context, request *proto.MoveRequest) (*proto.MoveResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggregate.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.MoveCMD{
			PlayerID: request.PlayerID,
			Position: request.Move,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error Move(%s)! details %s", request.GameID, err.Error(),
		)
	}

	return &proto.MoveResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) GetGame(ctx context.Context, request *proto.GetGameRequest) (*proto.GetGameResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggregate.Aggregate) error {
		return nil
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error GetGame(%s)! details %s", request.GameID, err.Error(),
		)
	}

	return &proto.GetGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func serialise(state *tictactoeaggregate.TicTacToeState) *proto.GameState {
	result := &proto.GameState{}

	if state.OneOf.GameWaitingForPlayer != nil {
		wg := state.OneOf.GameWaitingForPlayer
		result.State = &proto.GameState_Waiting{
			Waiting: &proto.GameState_GameWaitingForPlayer{
				NeedsPlayers: wg.NeedsPlayers,
			},
		}
	} else if state.OneOf.GameProgress != nil {
		gp := state.OneOf.GameProgress

		availableMoves := make([]string, 0)
		for move, _ := range gp.AvailableMoves {
			availableMoves = append(availableMoves, move)
		}

		result.State = &proto.GameState_Progress{
			Progress: &proto.GameState_GameProgress{
				NextMovePlayerID: gp.NextMovePlayerID,
				AvailableMoves:   availableMoves,
			},
		}
	} else if state.OneOf.GameResult != nil {
		gr := state.OneOf.GameResult
		result.State = &proto.GameState_Result{
			Result: &proto.GameState_GameResult{
				Winner:         gr.Winner,
				WiningSequence: gr.WiningSequence,
			},
		}
	} else {
		panic("Dragons! What state this is?")
	}

	return result
}
