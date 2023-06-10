package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
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

	game := di.GetGame()

	lambda.Start(func(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Infof("body: %#v \n", event.Body)
		err := game.OnMessage(event.RequestContext.ConnectionID, []byte(event.Body))
		if err != nil {
			log.Errorf("OnMessage: %s \n", err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
		}, nil
	})
}
