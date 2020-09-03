package awslambdabridge

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"reflect"
)

func NewAPIGatewayProxy() *Bridge {
	return &Bridge{
		before: make([]Middleware, 0),
	}
}

type Bridge struct {
	before []Middleware
	when   interface{}
	then   interface{}
}

type Middleware = func(ctx context.Context, request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse

func (b *Bridge) Use(m ...Middleware) *Bridge {
	b.before = append(b.before, m...)
	return b
}

type Request = func(ctx context.Context, request events.APIGatewayProxyRequest) (interface{}, error)

func (b *Bridge) When(f interface{}) *Bridge {
	b.when = f
	return b
}

type Response = func(ctx context.Context, response interface{}) (events.APIGatewayProxyResponse, error)

func (b *Bridge) Then(f interface{}) *Bridge {
	b.then = f
	return b
}

func (b *Bridge) Build() interface{} {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		for _, m := range b.before {
			response := m(ctx, request)
			if response != nil {
				return *response, nil
			}
		}

		input, err := callWhen(b.when, ctx, request)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		output := dispatch.Invoke(ctx, input)
		return callThen(b.then, ctx, output)
	}
}

func callWhen(f interface{}, ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(request),
	})

	return res[0].Interface().(events.APIGatewayProxyResponse), res[1].Interface().(error)
}

func callThen(f interface{}, ctx context.Context, output interface{}) (events.APIGatewayProxyResponse, error) {
	res := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(output),
	})

	return res[0].Interface().(events.APIGatewayProxyResponse), res[1].Interface().(error)
}
