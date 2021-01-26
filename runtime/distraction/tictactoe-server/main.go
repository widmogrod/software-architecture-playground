package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	eventstore "github.com/EventStore/EventStore-Client-Go/client"
	"github.com/EventStore/EventStore-Client-Go/direction"
	"github.com/EventStore/EventStore-Client-Go/messages"
	"github.com/EventStore/EventStore-Client-Go/streamrevision"
	uuid "github.com/gofrs/uuid"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/aggssert"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/prototictactoe"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/tictactoeaggregate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"math/rand"
	"net"
	"strings"
	"sync"
)

func main() {
	ln, _ := net.Listen("tcp", ":8080")
	server := grpc.NewServer()
	reflection.Register(server)

	var st dataStore

	if rand.Float64() < 0.9999 {
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

		//res, err := client.AppendToStream(context.Background(), "111aaaa", streamrevision.StreamRevisionNoStream, []messages.ProposedEvent{
		//	{EventID: uuid.Must(uuid.NewV4()), EventType:"gh-test", ContentType:"application/json", Data: []byte("{}"), UserMetadata:nil},
		//})
		//fmt.Println(res)
		//fmt.Println(err)
		//fmt.Printf("%#v\n", res.NextExpectedVersion)
		//fmt.Printf("%#v\n", streamrevision.NewStreamRevision(res.NextExpectedVersion))
		//res, err = client.AppendToStream(context.Background(), "111aaaa", streamrevision.NewStreamRevision(res.NextExpectedVersion), []messages.ProposedEvent{
		//	{EventID: uuid.Must(uuid.NewV4()), EventType:"gh-test", ContentType:"application/json", Data: []byte("{}"), UserMetadata:nil},
		//})
		//
		//fmt.Println(res)
		//fmt.Println(err)
		//fmt.Println(client.Connection.GetState().String())

		st = &evenstordbimpl{
			client: client,
		}
	} else {
		st = &inmemdatastor{
			data: sync.Map{},
		}
	}

	tictactoe := &tictactoeServer{
		store: &storere{
			new: func() aggssert.Aggregate {
				return tictactoeaggregate.NewTicTacToeAggregate()
			},
			store: st,
		},
	}
	prototictactoe.RegisterTicTacToeAggregateServer(server, tictactoe)
	_ = server.Serve(ln)
}

var _ prototictactoe.TicTacToeAggregateServer = &tictactoeServer{}

type tictactoeServer struct {
	store storer
}

var ErrNotFound = errors.New("AggregateID not found")

type dataStore interface {
	ReadChanges(ctx context.Context, aggregateID string) ([]runtime.Change, error)
	AppendChanges(ctx context.Context, aggregateID string, version uint64, changes []runtime.Change) error
}

type storer interface {
	NewAggregate(ctx context.Context, aggregateID string, handle handleFunc) (aggssert.Aggregate, error)
	MutateAggregate(ctx context.Context, aggregateID string, handle handleFunc) (aggssert.Aggregate, error)
}

type evenstordbimpl struct {
	client *eventstore.Client
}

func (i *evenstordbimpl) ReadChanges(ctx context.Context, aggregateID string) ([]runtime.Change, error) {
	// TODO read til exhaustion, don't wait for limit 20, or snapshot?
	events, err := i.client.ReadStreamEvents(ctx, direction.Forwards, aggregateID, streamrevision.StreamRevisionStart, 20, false)
	if err != nil {
		if strings.Contains(err.Error(), "stream was not found") {
			return nil, ErrNotFound
		}

		return nil, err
	}

	changes := make([]runtime.Change, 0)
	for _, event := range events {
		change := runtime.Change{
			Payload: nil,
			// TODO uint64!
			Version: event.EventNumber,
		}

		var err error
		switch event.EventType {
		case "GameCreated":
			change.Payload = &tictactoeaggregate.GameCreated{}
			err = json.Unmarshal(event.Data, change.Payload)
		case "SecondPlayerJoined":
			change.Payload = &tictactoeaggregate.SecondPlayerJoined{}
			err = json.Unmarshal(event.Data, change.Payload)
		case "Moved":
			change.Payload = &tictactoeaggregate.Moved{}
			err = json.Unmarshal(event.Data, change.Payload)
		case "GameFinish":
			change.Payload = &tictactoeaggregate.GameFinish{}
			err = json.Unmarshal(event.Data, change.Payload)
		default:
			err = fmt.Errorf("Not found undersialiser for type %s", event.EventType)
		}

		fmt.Printf("change:%#v, err=%w\n", change.Payload, err)
		if err != nil {
			return nil, fmt.Errorf("unnmarshall, %s - %w", event.EventType, err)
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (i *evenstordbimpl) AppendChanges(ctx context.Context, aggregateID string, version uint64, changes []runtime.Change) error {
	fmt.Println("AppendChanges ID=" + aggregateID)
	events := make([]messages.ProposedEvent, 0)

	for _, change := range changes {
		event := messages.ProposedEvent{
			ContentType: "application/json",
			EventID:     uuid.Must(uuid.DefaultGenerator.NewV4()),
		}

		var err error

		switch ch := change.Payload.(type) {
		case *tictactoeaggregate.GameCreated:
			event.EventType = "GameCreated"
			event.Data, err = json.Marshal(ch)
		case *tictactoeaggregate.SecondPlayerJoined:
			event.EventType = "SecondPlayerJoined"
			event.Data, err = json.Marshal(ch)
		case *tictactoeaggregate.Moved:
			event.EventType = "Moved"
			event.Data, err = json.Marshal(ch)
		case *tictactoeaggregate.GameFinish:
			event.EventType = "GameFinish"
			event.Data, err = json.Marshal(ch)
		default:
			err = fmt.Errorf("AppendChanges, Not found undersialiser for type %s", event.EventType)
		}

		if err != nil {
			return fmt.Errorf("AppendChanges, Marshall, %T - %w", change.Payload, err)
		}

		events = append(events, event)
	}

	//fmt.Printf("events=%#v\n", events)

	//res, err := i.client.AppendToStream(context.Background(), "111", streamrevision.StreamRevisionNoStream, []messages.ProposedEvent{
	//		{EventID: uuid.Must(uuid.NewV4()), EventType:"gh-test", ContentType:"application/json", Data: []byte("{}"), UserMetadata:nil},
	//	})

	var revision = streamrevision.StreamRevisionNoStream
	if version != ^uint64(0) {
		revision = streamrevision.NewStreamRevision(version)
	}

	fmt.Printf("version =%#v\n", version)
	fmt.Printf("revision =%#v\n", revision)
	res, err := i.client.AppendToStream(ctx, aggregateID, revision, events)
	fmt.Printf("res=%#v\n", res)
	fmt.Printf("err=%#v\n", err)
	if err != nil {
		return fmt.Errorf("AppendChanges %w", err)
	}

	return nil
}

type inmemdatastor struct {
	data sync.Map
}

func (i *inmemdatastor) ReadChanges(_ context.Context, aggregateID string) ([]runtime.Change, error) {
	data, found := i.data.Load(aggregateID)
	if !found {
		return nil, ErrNotFound
	}

	return data.([]runtime.Change), nil
}

func (i *inmemdatastor) AppendChanges(_ context.Context, aggregateID string, version uint64, changes []runtime.Change) error {
	i.data.Store(aggregateID, changes)
	return nil
}

type storere struct {
	new   newAgg
	store dataStore
}

type newAgg = func() aggssert.Aggregate

func (s *storere) NewAggregate(ctx context.Context, aggregateID string, handle handleFunc) (aggssert.Aggregate, error) {
	fmt.Println("NewAggregate ID=" + aggregateID)
	_, err := s.store.ReadChanges(ctx, aggregateID)
	if err == nil {
		return nil, fmt.Errorf("NewAggregate, on aggregate that exits %s", aggregateID)
	} else if err != ErrNotFound {
		return nil, fmt.Errorf("NewAggregate, unknow error on aggregate %s. Detail: %w", aggregateID, err)
	}

	agg := s.new()
	err = handle(agg)
	if err != nil {
		return nil, fmt.Errorf("NewAggregate, error while mutating aggregate %s. Details: %w", aggregateID, err)
	}

	err = s.save(ctx, aggregateID, ^uint64(0), agg)
	if err != nil {
		return nil, err
	}

	return agg, nil
}

type handleFunc = func(agg aggssert.Aggregate) error

func (s *storere) MutateAggregate(ctx context.Context, aggregateID string, handle handleFunc) (aggssert.Aggregate, error) {
	fmt.Println("MutateAggregate ID=" + aggregateID)
	var lastError error
	for retry := 0; retry < 2; retry++ {
		changes, err := s.store.ReadChanges(ctx, aggregateID)
		if err != nil {
			return nil, err
		}

		var version uint64 = 0
		agg := s.new()
		for _, change := range changes {
			version++
			err = agg.Changes().Append(change.Payload).Ok.ReduceRecent(agg).Err
			if err != nil {
				return nil, fmt.Errorf("MutateAggregate, error while replying aggregate %s, event=%#v. Details: %w", aggregateID, change, err)
			}
			version = change.Version
		}

		err = handle(agg)
		if err != nil {
			return nil, fmt.Errorf("MutateAggregate, error while mutating aggregate %s, event=%#v. Details: %w", aggregateID, err)
		}

		lastError = s.save(ctx, aggregateID, version, agg)
		if lastError == nil {
			return agg, nil
		}
	}

	return nil, fmt.Errorf("MutateAggregate, fail to store even after retrying X times. Details: %w", lastError)
}

func (s storere) save(ctx context.Context, aggregateID string, version uint64, agg aggssert.Aggregate) error {
	newChanges := make([]runtime.Change, 0)
	err := agg.Changes().ReduceChange(func(change runtime.Change, result *runtime.Reduced) *runtime.Reduced {
		if version == ^uint64(0) || change.Version > version {
			newChanges = append(newChanges, change)
		}
		return result
	}, nil).Err
	if err != nil {
		panic(fmt.Errorf("MutateAggregate, error while Reduce() changes must not happen. Details: %w", err))
	}

	return s.store.AppendChanges(ctx, aggregateID, version, newChanges)
}

func (s *tictactoeServer) CreateGame(ctx context.Context, request *prototictactoe.CreateGameRequest) (*prototictactoe.CreateGameResponse, error) {
	agg, err := s.store.NewAggregate(ctx, request.GameID, func(agg aggssert.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.CreateGameCMD{
			FirstPlayerID: request.FirstPlayerID,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error CreateGame()! details %w", request.GameID, err,
		)
	}

	return &prototictactoe.CreateGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) JoinGame(ctx context.Context, request *prototictactoe.JoinGameRequest) (*prototictactoe.JoinGameResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggssert.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.JoinGameCMD{
			SecondPlayerID: request.SecondPlayerID,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error JoinGame()! details %w", request.GameID, err,
		)
	}

	return &prototictactoe.JoinGameResponse{
		State: serialise(agg.State().(*tictactoeaggregate.TicTacToeState)),
	}, nil
}

func (s *tictactoeServer) Move(ctx context.Context, request *prototictactoe.MoveRequest) (*prototictactoe.MoveResponse, error) {
	agg, err := s.store.MutateAggregate(ctx, request.GameID, func(agg aggssert.Aggregate) error {
		return agg.Handle(&tictactoeaggregate.MoveCMD{
			PlayerID: request.PlayerID,
			Position: request.Move,
		})
	})
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Game error Move()! details %w", request.GameID, err,
		)
	}

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
