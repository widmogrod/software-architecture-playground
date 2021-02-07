package tictactoe

import (
	"flag"
	eventstore "github.com/EventStore/EventStore-Client-Go/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

var isIntegration = flag.Bool("i-exec-docker-compose-up", false, "Integration that tests require `docker-compose up`")

func TestEventStoreImplementationConformsToSpecification(t *testing.T) {
	if !*isIntegration {
		t.Skip("Skipping tests because this tests requires `docker-compose up`")
	}

	conf, err := eventstore.ParseConnectionString("esdb://admin:changeit@localhost:2113?tls=false&tlsverifycert=false")
	assert.NoError(t, err)

	client, err := eventstore.NewClient(conf)
	assert.NoError(t, err)

	_ = client.Connect()
}
