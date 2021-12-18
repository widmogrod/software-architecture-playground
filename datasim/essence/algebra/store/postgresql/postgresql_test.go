package postgresql

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra/store"
	"testing"
)

var isIntegration = flag.Bool("i-exec-docker-compose-up", false, "Integration that tests require `docker-compose up`")

var (
	relation = store.PrimaryWithMany{
		Primary: store.Entity{
			Name: "question",
			Attributes: []store.Attr{
				store.Attribute{Name: "authorId", Type: store.String{}},
				store.Attribute{Name: "content", Type: store.String{}},
			},
		},
		Secondaries: []store.Shape{
			store.Entity{
				Name: "quality",
				Attributes: []store.Attr{
					store.Attribute{Name: "toxic_score", Type: store.Int{}},
					store.Attribute{Name: "sentiment_score", Type: store.Int{}},
					store.Attribute{Name: "quality_score", Type: store.Int{}},
				},
			},
			store.Entity{
				Name: "answer",
				Attributes: []store.Attr{
					store.Attribute{Name: "authorId", Type: store.String{}},
					store.Attribute{Name: "content", Type: store.String{}},
					store.Attribute{Name: "createdAt", Type: store.DateTime{}},
				},
			},
			store.Entity{
				Name: "comment",
				Attributes: []store.Attr{
					store.Attribute{Name: "authorId", Type: store.String{}},
					store.Attribute{Name: "content", Type: store.String{}},
					store.Attribute{Name: "createdAt", Type: store.DateTime{}},
				},
			},
		},
	}
)

func TestInitiateStoreShape(t *testing.T) {
	//if !*isIntegration {
	//	t.Skip("Skipping tests because this tests requires `docker-compose up`")
	//}

	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("cannot establish database connection: %s", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("cannot ping database: %s", err)
	}

	//cleanUpDatabase(t, db)

	s := NewStore(db, relation)
	err = s.InitiateShape()
	assert.NoError(t, err)

	err = s.Set("1", "answer", "content", "asd")
	assert.NoError(t, err)

	res, err := s.Get("1", "answer", store.AttrList{"content"})
	assert.NoError(t, err)

	fmt.Printf("res=%#v", res)
}

func failOn(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("coudn't complete action: %s", err)
	}
}
