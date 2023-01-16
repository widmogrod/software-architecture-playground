package websockproto

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"net/url"
)

var _ Publisher = (*AWSPublisher)(nil)

type AWSPublisher struct {
	DomainName string
	Stage      string
	cfg        aws.Config
}

func (a *AWSPublisher) Publish(connectionID string, msg []byte) error {
	callbackURL := url.URL{
		Scheme: "https",
		Host:   a.DomainName,
		Path:   a.Stage,
	}

	fmt.Printf("callbackURL: %s \n", callbackURL.String())

	api := apigatewaymanagementapi.NewFromConfig(
		a.cfg,
		apigatewaymanagementapi.WithEndpointResolver(
			apigatewaymanagementapi.EndpointResolverFromURL(
				callbackURL.String(),
			),
		),
	)

	_, err := api.PostToConnection(context.Background(), &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         msg,
	})

	if err != nil {
		return err
	}
	return nil
}
