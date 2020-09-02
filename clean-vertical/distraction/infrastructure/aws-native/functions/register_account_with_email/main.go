package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation/inmemory"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

func requestToCommand(request events.APIGatewayProxyRequest) (usecase.RegisterAccountWithEmail, error) {
	command := usecase.RegisterAccountWithEmail{}
	if err := json.Unmarshal([]byte(request.Body), &command); err != nil {
		return command, err
	}
	return command, nil
}

func resultToResponse(output usecase.ResultOfRegisteringWithEmail) (events.APIGatewayProxyResponse, error) {
	if output.IsSuccessful() {
		body, err := json.Marshal(output.SuccessfulResult)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       err.Error(),
			}, err
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(body),
		}, nil
	}

	body, err := json.Marshal(output.ValidationError)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       string(body),
	}, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println(request.Body)
	cmd, err := requestToCommand(request)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	res := dispatch.Invoke(ctx, cmd)
	return resultToResponse(res.(usecase.ResultOfRegisteringWithEmail))
}

func main() {
	interpretation := inmemory.New()
	dispatch.Interpret(interpretation)

	// boot middleware that handle JWT transformation
	// do permission extraction
	// do it with dispatch O.o
	lambda.Start(handler)
}
