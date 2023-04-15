package tictactoe_game_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/typedful"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"testing"
	"time"
)

func TestNewLiveSelect(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "",
		PadLevelText:    true,
	})

	store := schemaless.NewInMemoryRepository()
	stream := store.AppendLog()

	typedStore := typedful.NewTypedRepository[tictactoemanage.State](store)
	broadcast := &BroadcasterMock{
		BroadcastToSessionFunc: func(sessionID string, msg []byte) {},
	}

	update := schemaless.UpdateRecords[schemaless.Record[tictactoemanage.State]]{
		Saving: map[string]schemaless.Record[tictactoemanage.State]{},
	}
	for _, game := range latestGames {
		id := tictactoemanage.MustMatchState(
			game,
			func(x *tictactoemanage.SessionWaitingForPlayers) string {
				return x.SessionID
			},
			func(x *tictactoemanage.SessionReady) string {
				return x.SessionID
			},
			func(x *tictactoemanage.SessionInGame) string {
				return x.SessionID + "-" + x.GameID
			},
		)
		update.Saving["game:"+id] = schemaless.Record[tictactoemanage.State]{
			ID:      id,
			Type:    "game",
			Data:    game,
			Version: 1,
		}
		update.Saving["session:"+id] = schemaless.Record[tictactoemanage.State]{
			ID:      id,
			Type:    "session",
			Data:    game,
			Version: 1,
		}
	}

	err := typedStore.UpdateRecords(update)
	assert.NoError(t, err)

	// send signal for DAG to not wait for any future changes
	// there won't be any.
	stream.Close()

	ctx := context.Background()
	live := NewLiveSelect(typedStore, broadcast)

	go func() {
		<-time.After(1 * time.Second)
		live.UseStreamToPush(stream)
	}()

	err = live.Process(ctx, "session-1")
	assert.NoError(t, err)

	assert.Len(t, broadcast.BroadcastToSessionCalls(), 3)

	// last state should look like this:
	assert.EqualValues(t, struct {
		SessionID string
		Msg       []byte
	}{
		SessionID: "session-1",
		Msg:       []byte(`{"SessionStatsResult":{"ID":"session-1","TotalGames":3.000000,"TotalDraws":1.000000,"PlayerWins":{"player-1":2.000000}}}`),
	}, broadcast.BroadcastToSessionCalls()[2])

	//err = live.Process(ctx, "session-1")
	//assert.NoError(t, err)
	//
	//assert.Len(t, broadcast.BroadcastToSessionCalls(), 6)
}
