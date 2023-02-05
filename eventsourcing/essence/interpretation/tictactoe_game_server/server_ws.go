package tictactoe_game_server

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

func NewWebSocket(ctx context.Context) (*websockproto.InMemoryProtocol, error) {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	connRepo := storage.NewDynamoDBRepository(
		dynamodb.NewFromConfig(cfg),
		"test-repo",
		func() websockproto.ConnectionToSession {
			panic("not supported creation of ConnectionToSession")
		})

	stateRepo := storage.NewDynamoDBRepository(
		dynamodb.NewFromConfig(cfg),
		"test-repo",
		func() tictactoemanage.State {
			return nil
		})

	wshandler := websockproto.NewInMemoryProtocol()
	broadcaster := websockproto.NewBroadcaster(wshandler, connRepo)

	game := &Game{
		broadcast:           broadcaster,
		gameStateRepository: stateRepo,
	}

	wshandler.OnMessage = game.OnMessage
	wshandler.OnConnect = game.OnConnect
	wshandler.OnDisconnect = game.OnDisconnect

	return wshandler, nil
}
