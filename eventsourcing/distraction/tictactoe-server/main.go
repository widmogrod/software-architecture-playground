package main

import (
	"context"
	eventstore "github.com/EventStore/EventStore-Client-Go/client"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate/store/eventstoredb"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/deserializer"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/prototictactoe"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/tictactoeaggregate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
)

func main() {
	ln, _ := net.Listen("tcp", ":8080")
	server := grpc.NewServer()
	reflection.Register(server)

	conf, err := eventstore.ParseConnectionString("esdb://admin:changeit@localhost:2113?tls=false&tlsverifycert=false")
	if err != nil {
		panic(err)
	}

	client, err := eventstore.NewClient(conf)
	if err != nil {
		panic(err)
	}

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	ser := deserializer.NewDeSerializer()
	ser.Register(tictactoeaggregate.GameFinish{})
	ser.Register(tictactoeaggregate.Moved{})
	ser.Register(tictactoeaggregate.SecondPlayerJoined{})
	ser.Register(tictactoeaggregate.GameCreated{})

	st := eventstoredb.NewEventStoreDB(client, ser)

	tictactoe := &tictactoeServer{
		store: aggregate.NewAggregate(func() aggregate.Aggregate {
			return tictactoeaggregate.NewTicTacToeAggregate()
		}, st),
	}
	prototictactoe.RegisterTicTacToeAggregateServer(server, tictactoe)
	_ = server.Serve(ln)
}

var _ prototictactoe.TicTacToeAggregateServer = &tictactoeServer{}

type storer interface {
	NewAggregate(ctx context.Context, aggregateID string, handle aggregate.HandleFunc) (aggregate.Aggregate, error)
	MutateAggregate(ctx context.Context, aggregateID string, handle aggregate.HandleFunc) (aggregate.Aggregate, error)
}

type tictactoeServer struct {
	store storer
}

func (s *tictactoeServer) CreateGame(ctx context.Context, request *prototictactoe.CreateGameRequest) (*prototictactoe.CreateGameResponse, error) {
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

	return &prototictactoe.CreateGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) JoinGame(ctx context.Context, request *prototictactoe.JoinGameRequest) (*prototictactoe.JoinGameResponse, error) {
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

	return &prototictactoe.JoinGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) Move(ctx context.Context, request *prototictactoe.MoveRequest) (*prototictactoe.MoveResponse, error) {
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

	return &prototictactoe.MoveResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) GetGame(ctx context.Context, request *prototictactoe.GetGameRequest) (*prototictactoe.GetGameResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggregate.Aggregate) error {
		return nil
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error GetGame(%s)! details %s", request.GameID, err.Error(),
		)
	}

	return &prototictactoe.GetGameResponse{
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
