package websockproto

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"log"
	"net/url"
)

func NewPublisher(domainName, stage string, cfg aws.Config) *AWSPublisher {
	callbackURL := url.URL{
		Scheme: "https",
		Host:   domainName,
		Path:   stage,
	}

	api := apigatewaymanagementapi.NewFromConfig(
		cfg,
		apigatewaymanagementapi.WithEndpointResolver(
			apigatewaymanagementapi.EndpointResolverFromURL(
				callbackURL.String(),
			),
		),
	)
	return &AWSPublisher{
		client: api,
	}
}

var _ Publisher = (*AWSPublisher)(nil)

type AWSPublisher struct {
	client *apigatewaymanagementapi.Client
}

func (a *AWSPublisher) Publish(connectionID string, msg []byte) error {
	log.Println("Publishing to connectionID:", connectionID, "msg:", string(msg))
	_, err := a.client.PostToConnection(context.Background(), &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         msg,
	})

	if err != nil {
		return err
	}
	return nil
}
