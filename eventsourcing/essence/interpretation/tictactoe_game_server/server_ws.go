package tictactoe_game_server

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
)

func NewWebSocket(ctx context.Context) (*websockproto.InMemoryProtocol, error) {
	di := DefaultDI(RunLocalInMemoryDevelopment)

	game := di.GetGame()
	wshandler := di.GetInMemoryWebSocketPublisher()
	wshandler.OnMessage = game.OnMessage
	wshandler.OnConnect = game.OnConnect
	wshandler.OnDisconnect = game.OnDisconnect

	return wshandler, nil
}
