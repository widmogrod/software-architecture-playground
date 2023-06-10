package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"net/url"
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
		for i := range event.Records {
			record := event.Records[i]
			payload := Payload{}
			err := json.Unmarshal([]byte(record.Body), &payload)
			if err != nil {
				fmt.Printf("ERR: unmarshall: %s \n", err)
				continue
			}

			callbackURL := url.URL{
				Scheme: "https",
				Host:   payload.RequestContext.DomainName,
				Path:   payload.RequestContext.Stage,
			}

			fmt.Printf("callbackURL: %s \n", callbackURL.String())

			a := apigatewaymanagementapi.NewFromConfig(
				cfg,
				apigatewaymanagementapi.WithEndpointResolver(
					apigatewaymanagementapi.EndpointResolverFromURL(
						callbackURL.String(),
					),
				),
			)

			_, err = a.PostToConnection(context.Background(), &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: &payload.ConnectionID,
				Data:         []byte("go:" + payload.Body),
			})

			if err != nil {
				fmt.Printf("ERR: PostToConnection: %s \n", err)
				continue
			}
		}
	})
}
