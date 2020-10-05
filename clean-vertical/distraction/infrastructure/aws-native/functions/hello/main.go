package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/awslambdabridge"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation/inmemory"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

func main() {
	interpretation := inmemory.New()
	dispatch.Interpret(interpretation)

	handler := awslambdabridge.NewAPIGatewayProxy().
		When(func(ctx context.Context, request events.APIGatewayProxyRequest) (usecase.HelloWorld, error) {
			return usecase.HelloWorld{
				Name: request.QueryStringParameters["name"],
			}, nil
		}).
		Then(func(ctx context.Context, output usecase.ResultOfHelloWorld) (events.APIGatewayProxyResponse, error) {
			body, err := json.Marshal(output)
			if err != nil {
				return events.APIGatewayProxyResponse{}, err
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(body),
			}, nil
		}).
		Build()

	lambda.Start(handler)
}
