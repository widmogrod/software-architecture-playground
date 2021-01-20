package main

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/prototictactoe"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/tictactoeaggregate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
	"sync"
)

func main() {
	ln, _ := net.Listen("tcp", ":8080")
	server := grpc.NewServer()
	reflection.Register(server)

	tictactoe := &tictactoeServer{
		store: sync.Map{},
	}
	prototictactoe.RegisterTicTacToeAggregateServer(server, tictactoe)
	_ = server.Serve(ln)
}

var _ prototictactoe.TicTacToeAggregateServer = &tictactoeServer{}

type tictactoeServer struct {
	store sync.Map
}

func (s *tictactoeServer) CreateGame(ctx context.Context, request *prototictactoe.CreateGameRequest) (*prototictactoe.CreateGameResponse, error) {
	if _, ok := s.store.Load(request.GameID); ok {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"Game already exists id = ", request.GameID,
		)
	}

	agg := tictactoeaggregate.NewTicTacToeAggregate()
	err := agg.Handle(&tictactoeaggregate.CreateGameCMD{
		FirstPlayerID: request.FirstPlayerID,
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"err agg.Handle(CreateGameCMD{}) = %s", err,
		)
	}

	s.store.Store(request.GameID, agg)

	return &prototictactoe.CreateGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) JoinGame(ctx context.Context, request *prototictactoe.JoinGameRequest) (*prototictactoe.JoinGameResponse, error) {
	res, found := s.store.Load(request.GameID)
	if !found {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"Game don't exists id = ", request.GameID,
		)
	}

	agg := res.(*tictactoeaggregate.TicTacToeAggregate)
	err := agg.Handle(&tictactoeaggregate.JoinGameCMD{
		SecondPlayerID: request.SecondPlayerID,
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"err agg.Handle(JoinGameCMD{}) = %s", err,
		)
	}

	s.store.Store(request.GameID, agg)

	return &prototictactoe.JoinGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) Move(ctx context.Context, request *prototictactoe.MoveRequest) (*prototictactoe.MoveResponse, error) {
	res, found := s.store.Load(request.GameID)
	if !found {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"Game don't exists id = ", request.GameID,
		)
	}

	agg := res.(*tictactoeaggregate.TicTacToeAggregate)
	err := agg.Handle(&tictactoeaggregate.MoveCMD{
		PlayerID: request.PlayerID,
		Position: request.Move,
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"err agg.Handle(MoveCMD{}) = %s", err,
		)
	}

	s.store.Store(request.GameID, agg)

	return &prototictactoe.MoveResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func serialise(state *tictactoeaggregate.TicTacToeState) *prototictactoe.GameState {
	result := &prototictactoe.GameState{}

	if state.OneOf.GameWaitingForPlayer != nil {
		wg := state.OneOf.GameWaitingForPlayer
		result.State = &prototictactoe.GameState_Waiting{
			Waiting: &prototictactoe.GameState_GameWaitingForPlayer{
				NeedsPlayers: wg.NeedsPlayers,
			},
		}
	} else if state.OneOf.GameProgress != nil {
		gp := state.OneOf.GameProgress

		availableMoves := make([]string, 0)
		for move, _ := range gp.AvailableMoves {
			availableMoves = append(availableMoves, move)
		}

		result.State = &prototictactoe.GameState_Progress{
			Progress: &prototictactoe.GameState_GameProgress{
				NextMovePlayerID: gp.NextMovePlayerID,
				AvailableMoves:   availableMoves,
			},
		}
	} else if state.OneOf.GameResult != nil {
		gr := state.OneOf.GameResult
		result.State = &prototictactoe.GameState_Result{
			Result: &prototictactoe.GameState_GameResult{
				Winner:         gr.Winner,
				WiningSequence: gr.WiningSequence,
			},
		}
	} else {
		panic("Dragons! What state this is?")
	}

	return result
}
