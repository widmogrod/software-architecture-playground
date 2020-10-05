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
	"net/http"
	"strings"
)

type key struct{}

var identityKey = &key{}

type Identity struct {
	roles  []string
	userID int
}

func (i Identity) Roles() []string {
	return i.roles
}

func authjwt(ctx context.Context, request events.APIGatewayProxyRequest) (context.Context, *events.APIGatewayProxyResponse) {
	bearToken := request.Headers["Authorization"]
	parts := strings.Split(bearToken, " ")
	if len(parts) != 2 {
		return ctx, &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
		}
	}

	// TODO JWT validation & claims extraction
	_ = parts[1]
	claims := map[string]interface{}{
		"roles": []string{"guest_user"},
		"uid":   123,
	}

	identity := Identity{
		roles:  claims["roles"].([]string),
		userID: claims["uid"].(int),
	}

	return context.WithValue(ctx, identityKey, identity), nil
}

type Permissions struct {
	rolesAccess map[string]map[string]struct{}
}

func (p Permissions) Can(identity interface{ Roles() []string }, do string) bool {
	for _, role := range identity.Roles() {
		if cans, ok := p.rolesAccess[role]; ok {
			if _, ok := cans[do]; ok {
				return true
			}
		}
	}

	return false
}

func allowAccess(perm interface {
	Can(identity interface{ Roles() []string }, do string) bool
}, access string) awslambdabridge.Middleware {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (context.Context, *events.APIGatewayProxyResponse) {
		identity, ok := ctx.Value(identityKey).(Identity)
		if !ok {
			return ctx, &events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
				Body:       "no identity",
			}
		}

		if !perm.Can(identity, access) {
			return ctx, &events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
				Body:       "no can do",
			}
		}

		return ctx, nil
	}
}

func main() {
	perm := Permissions{
		rolesAccess: map[string]map[string]struct{}{
			"guest_user": {
				"can_register": struct{}{},
			},
		},
	}

	interpretation := inmemory.New()
	dispatch.Interpret(interpretation)

	handler := awslambdabridge.NewAPIGatewayProxy().
		Use(authjwt).
		Use(allowAccess(perm, "can_register")).
		When(func(ctx context.Context, request events.APIGatewayProxyRequest) (usecase.RegisterAccountWithEmail, error) {
			command := usecase.RegisterAccountWithEmail{}
			if err := json.Unmarshal([]byte(request.Body), &command); err != nil {
				return command, err
			}
			return command, nil
		}).
		Then(func(ctx context.Context, output usecase.ResultOfRegisteringWithEmail) (events.APIGatewayProxyResponse, error) {
			body, err := json.Marshal(output)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
					Body:       err.Error(),
				}, err
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(body),
			}, err
		}).
		Build()

	lambda.Start(handler)
}
