package tictactoe_game_server

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"testing"
)

func TestQuery(t *testing.T) {
	q, err := NewQuery(
		"http://localhost:9200",
		"lambda-index",
	)
	assert.NoError(t, err)

	result, err := q.Query(tictactoemanage.SessionStatsQuery{SessionID: "605e54ac-1d84-4ccf-9004-df4a21c98d5f"})
	assert.NoError(t, err)

	t.Logf("res: \n\t %#v \n", result)
}
