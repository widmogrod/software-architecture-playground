package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      false,
		DisableQuote:     true,
		DisableTimestamp: true,
	})

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunAWS,
	)

	lambda.Start(&handler{
		liveSelectClient: di.GetLiveSelectClient(),
	})
}

type handler struct {
	liveSelectClient *tictactoe_game_server.LiveSelectClient
}

func (h handler) Invoke(ctx context.Context, data []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("Received event: panic", r)
		}
	}()

	log.Debugln("Received event: ", string(data))
	err := h.liveSelectClient.Push(ctx, data)
	if err != nil {
		log.Error("Received event: ERR", err)
	}

	return []byte(nil), nil
}
