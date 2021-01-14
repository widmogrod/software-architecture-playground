package interpretation

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

	conf, err := eventstore.ParseConnectionString("tcp://localhost:1113")
	assert.NoError(t, err)

	client, err := eventstore.NewClient(conf)
	assert.NoError(t, err)

	_ = client.Connect()
}
