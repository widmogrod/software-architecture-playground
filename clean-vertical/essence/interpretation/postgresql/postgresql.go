package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"math/rand"
	"strconv"
)

var _ interpretation.Interpretation = &PostgresSQL{}

func New(db *sql.DB) *PostgresSQL {
	return &PostgresSQL{
		db: db,
	}
}

type PostgresSQL struct {
	db *sql.DB
}

func (e *PostgresSQL) HandleCreateUserIdentity(ctx dispatch.Context, input usecase.CreateUserIdentity) usecase.ResultOfCreateUserIdentity {
	output := usecase.ResultOfCreateUserIdentity{}

	_, err := e.db.ExecContext(ctx.Ctx(), "INSERT INTO user_identity(uuid, email_address) VALUES($1, $2)", input.UUID, input.EmailAddress)
	if err != nil {
		output.ValidationError = usecase.NewConflictEmailExistsError()
	} else {
		output.SuccessfulResult = usecase.NewCreateUserIdentityWithUUID(input.UUID)
	}

	return output
}

func (e *PostgresSQL) HandleGenerateSessionToken(ctx dispatch.Context, input usecase.GenerateSessionToken) usecase.ResultOfGeneratingSessionToken {
	output := usecase.ResultOfGeneratingSessionToken{
		SuccessfulResult: usecase.SessionToken{
			AccessToken:  strconv.Itoa(rand.Int()),
			RefreshToken: strconv.Itoa(rand.Int()),
		},
	}
	return output
}

type ActivationTokenEntity struct {
	Token string
	UUID  string
}

func (e *PostgresSQL) HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input usecase.MarkAccountActivationTokenAsUse) usecase.ResultOfMarkingAccountActivationTokenAsUsed {
	row := e.db.QueryRowContext(ctx.Ctx(), "DELETE FROM activation_tokens WHERE token=$1 RETURNING uuid", input.ActivationToken)

	output := usecase.ResultOfMarkingAccountActivationTokenAsUsed{}
	if row.Err() != nil {
		panic("query row did not work! " + row.Err().Error())
	} else {
		token := ""
		err := row.Scan(&token)
		if err != nil {
			if err == sql.ErrNoRows {
				output.ValidationError = usecase.NewAccountActivationInvalidTokenError()
			} else {
				panic("row scan did not work! " + err.Error())
			}
		} else {
			output.SuccessfulResult = usecase.NewAccountActivatedViaTokenSuccess(token)
		}
	}

	return output
}

func (e *PostgresSQL) HandleCreateAccountActivationToken(ctx dispatch.Context, input usecase.CreateAccountActivationToken) usecase.ResultOfCreateAccountActivationToken {
	token := ksuid.New().String()

	_, err := e.db.ExecContext(ctx.Ctx(), "INSERT INTO activation_tokens(uuid, token) VALUES($1, $2)", input.UUID, token)

	if err != nil {
		panic("cannot creat token! " + err.Error())
	}

	output := usecase.ResultOfCreateAccountActivationToken{}
	output.SuccessfulResult = token
	return output
}

func (e *PostgresSQL) HandleHelloWorld(ctx dispatch.Context, input usecase.HelloWorld) usecase.ResultOfHelloWorld {
	return usecase.ResultOfHelloWorld{
		SuccessfulResult: fmt.Sprintf("Ola! %s", input),
	}
}
