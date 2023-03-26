package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
)

type Payload struct {
	RequestContext events.APIGatewayWebsocketProxyRequestContext `json:"requestContext"`
	ConnectionID   string                                        `json:"connectionId"`
	Body           string                                        `json:"body"`
}

func main() {
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "",
		PadLevelText:    true,
	})

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunAWS,
	)

	game := di.GetGame()

	lambda.Start(func(ctx context.Context, event events.SQSEvent) {
		for i := range event.Records {
			record := event.Records[i]
			payload := Payload{}
			err := json.Unmarshal([]byte(record.Body), &payload)
			if err != nil {
				fmt.Printf("ERR: Unmarshal: %s \n", err)
				continue
			}

			log.Printf("payload: %#v \n", payload)

			err = game.OnMessage(payload.ConnectionID, []byte(payload.Body))
			if err != nil {
				fmt.Printf("ERR: OnMessage: %s \n", err)
				continue
			}
		}
	})
}
