package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"log"
	"os"
)

type Payload struct {
	RequestContext events.APIGatewayWebsocketProxyRequestContext `json:"requestContext"`
	ConnectionID   string                                        `json:"connectionId"`
	Body           string                                        `json:"body"`
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Printf("ERR: LoadDefaultConfig: %s \n", err)
		panic(err)
	}

	lambda.Start(func(ctx context.Context, event events.SQSEvent) {
		tableName := os.Getenv("TABLE_NAME")
		log.Println("TABLE_NAME: ", tableName)

		for i := range event.Records {
			record := event.Records[i]
			payload := Payload{}
			err := json.Unmarshal([]byte(record.Body), &payload)
			if err != nil {
				fmt.Printf("ERR: Unmarshal: %s \n", err)
				continue
			}

			log.Println("payload: ", payload)

			connRepo := storage.NewDynamoDBRepository(
				dynamodb.NewFromConfig(cfg),
				tableName,
				func() websockproto.ConnectionToSession {
					panic("not supported creation of ConnectionToSession")
				})

			stateRepo := storage.NewDynamoDBRepository(
				dynamodb.NewFromConfig(cfg),
				tableName,
				func() tictactoemanage.State {
					return nil
				})

			wshandler := websockproto.NewPublisher(
				payload.RequestContext.DomainName,
				payload.RequestContext.Stage,
				cfg,
			)
			broadcaster := websockproto.NewBroadcaster(wshandler, connRepo)

			game := tictactoe_game_server.NewGame(broadcaster, stateRepo)
			err = game.OnMessage(payload.ConnectionID, []byte(payload.Body))
			if err != nil {
				fmt.Printf("ERR: OnMessage: %s \n", err)
				continue
			}
		}
	})
}
