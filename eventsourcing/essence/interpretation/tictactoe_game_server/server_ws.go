package tictactoe_game_server

import (
	"context"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

func NewWebSocket(ctx context.Context) (*websockproto.InMemoryProtocol, error) {
	var err error
	//cfg, err := config.LoadDefaultConfig(ctx)
	//if err != nil {
	//	return nil, err
	//}

	store := storage.NewRepository2WithSchema()
	connRepo := storage.NewRepository2Typed[websockproto.ConnectionToSession](store)
	//connRepo := storage.NewDynamoDBRepository(
	//	dynamodb.NewFromConfig(cfg),
	//	"test-repo",
	//	func() websockproto.ConnectionToSession {
	//		panic("not supported creation of ConnectionToSession")
	//	})

	stateRepo := storage.NewRepositoryWithAggregator[tictactoemanage.State, tictactoemanage.SessionStatsResult](
		store,
		func() storage.Aggregator[tictactoemanage.State, tictactoemanage.SessionStatsResult] {
			return NewTictactoeManageStateAggregate(store)
		},
	)
	//stateRepo := storage.NewDynamoDBRepository(
	//	dynamodb.NewFromConfig(cfg),
	//	"test-repo",
	//	func() tictactoemanage.State {
	//		return nil
	//	})

	query := NewQueryUsingStorage(
		storage.NewRepository2Typed[tictactoemanage.SessionStatsResult](store),
	)
	//query, err := NewQuery(
	//	"https://search-dynamodb-projection-vggyq7lvwooliwe65oddc5gyse.eu-west-1.es.amazonaws.com/",
	//	"lambda-index",
	//)
	if err != nil {
		fmt.Printf("ERR: tictactoe_game_server.NewQuery: %s \n", err)
		panic(err)
	}

	wshandler := websockproto.NewInMemoryProtocol()
	broadcaster := websockproto.NewBroadcaster(wshandler, connRepo)

	game := &Game{
		broadcast:           broadcaster,
		gameStateRepository: stateRepo,
		query:               query,
	}

	wshandler.OnMessage = game.OnMessage
	wshandler.OnConnect = game.OnConnect
	wshandler.OnDisconnect = game.OnDisconnect

	return wshandler, nil
}
