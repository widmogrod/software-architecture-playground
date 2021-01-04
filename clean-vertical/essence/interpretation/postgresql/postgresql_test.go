package postgresql

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	"testing"
)

var isIntegration = flag.Bool("i-exec-docker-compose-up", false, "Integration that tests require `docker-compose up`")

func TestPostgreSQLImplementationConformsToSpecification(t *testing.T) {
	if !*isIntegration {
		t.Skip("Skipping tests because this tests requires `docker-compose up`")
	}

	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("cannot establish database connection: %s", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("cannot ping database: %s", err)
	}

	cleanUpDatabase(t, db)

	implementation := New(db)
	dispatch.Interpret(implementation)
	interpretation.Specification(t)
}

func cleanUpDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE TABLE activation_tokens`)
	failOn(t, err)
	_, err = db.Exec(`TRUNCATE TABLE user_identity`)
	failOn(t, err)
}

func failOn(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("coudn't complete action: %s", err)
	}
}
