package tictactoe_game_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/projection"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"sync"
)

//go:generate moq -out live_select_broadcaster_moq_test.go . Broadcaster

type Broadcaster interface {
	BroadcastToSession(sessionID string, msg []byte)
}

type Stream interface {
	Subscribe(ctx context.Context, fromOffset int, f func(change schemaless.Change[schema.Schema])) error
}

type LiveSelect struct {
	store     schemaless.Repository[tictactoemanage.State]
	broadcast Broadcaster
	pubsub    *projection.PubSubChan[schemaless.Change[schema.Schema]]
	once      sync.Once
}

func NewLiveSelect(
	store schemaless.Repository[tictactoemanage.State],
	broadcast Broadcaster,
) *LiveSelect {
	return &LiveSelect{
		store:     store,
		broadcast: broadcast,
		pubsub:    projection.NewPubSubChan[schemaless.Change[schema.Schema]](),
	}
}

func (l *LiveSelect) Push(ctx context.Context, change schemaless.Change[schema.Schema]) error {
	log.Infoln("LiveSelect.Push", "change", change)
	l.once.Do(func() {
		go l.pubsub.Process()
	})
	return l.pubsub.Publish(change)
}

func (l *LiveSelect) UseStreamToPush(stream Stream) *LiveSelect {
	go func() {
		defer l.pubsub.Close()

		ctx := context.Background()
		err := stream.Subscribe(ctx, 0, func(change schemaless.Change[schema.Schema]) {
			err := l.Push(ctx, change)
			if err != nil {
				log.WithError(err).Error("failed to push change to live select")
			}
		})
		if err != nil {
			log.WithError(err).Error("failed to subscribe to stream")
		}
	}()

	return l
}

func (l *LiveSelect) Process(ctx context.Context, sessionID string) error {
	where := predicate.MustWhere(
		"Data.SessionInGame.SessionID = :sessionID AND Type = :type",
		map[predicate.BindValue]schema.Schema{
			":sessionID": schema.FromGo(sessionID),
			":type":      schema.FromGo("game"),
		},
	)

	dag := projection.NewDAGBuilder()
	// Only latest records from database that match live select criteria are used
	lastState := dag.
		Load(&projection.GenerateHandler{
			Load: func(push func(message projection.Item)) error {
				results, err := l.store.FindingRecords(schemaless.FindingRecords[schemaless.Record[tictactoemanage.State]]{
					Where: where,
				})

				if err != nil {
					return err
				}

				for _, item := range results.Items {
					push(projection.Item{
						Key:  item.ID,
						Data: l.fromTyped(item),
					})
				}

				return nil
			},
		}, projection.WithName("1. Load DynamoDB LastState Filtered"))
	// Only streamed records that match live select criteria are used
	streamState := dag.
		Load(&projection.GenerateHandler{
			Load: func(push func(message projection.Item)) error {
				log.Infoln("LiveSelect.Process", "subscribing to pubsub")
				defer log.Infoln("LiveSelect.Process", "subscribing to pubsub END")
				return l.pubsub.Subscribe(func(change schemaless.Change[schema.Schema]) error {
					log.Infoln("LiveSelect.Process received", "change", change)
					if change.Deleted {
						log.Warnf("Item was deleted: %v, live select skip on it", change)
						return nil
					}

					record := *change.After

					log.Infoln("LiveSelect.Process pushed", "change", change)
					push(projection.Item{
						Key:  record.ID,
						Data: l.fromUnTyped(record),
					})
					return nil
				})
			},
		}, projection.WithName("2. Load changes from channel"))
	// Joining make sure that newest version is published

	joined := dag.
		// Join by key, so if db and stream has the same key, then it will be joined.
		Join(lastState, streamState,
			projection.WithName("3. Join [DB & Stream]")).
		Map(&projection.FilterHandler{
			Where: where,
		}, projection.WithName("4. Filter[Join [DB & Stream]]")).
		// Joining by key and producing a new key is like merging!
		Merge(&projection.JoinHandler[schemaless.Record[tictactoemanage.State]]{
			F: func(a, b schemaless.Record[tictactoemanage.State], returning func(schemaless.Record[tictactoemanage.State])) error {
				if a.Version < b.Version {
					returning(b)
				}
				return nil
			},
		}, projection.WithName("5. Merge (version)[Join [DB & Stream]]"))

	gameStats := joined.
		Map(&projection.MapHandler[schemaless.Record[tictactoemanage.State], tictactoemanage.SessionStatsResult]{
			F: func(x schemaless.Record[tictactoemanage.State], returning func(key string, value tictactoemanage.SessionStatsResult)) error {
				returning(GroupByKey(x.Data))
				return nil
			},
		}, projection.WithName("6. Map GameToStats"))

	gameStatsBySession := gameStats.
		Merge(&projection.MergeHandler[tictactoemanage.SessionStatsResult]{
			Combine: CombineByKey,
			DoRetract: func(base tictactoemanage.SessionStatsResult, x tictactoemanage.SessionStatsResult) (tictactoemanage.SessionStatsResult, error) {
				panic("retract not implemented")
			},
		}, projection.WithName("7. Merge SessionStats"))

	gameStatsBySession.
		Map(&projection.MapHandler[tictactoemanage.SessionStatsResult, any]{
			F: func(x tictactoemanage.SessionStatsResult, returning func(key string, value any)) error {
				var r tictactoemanage.QueryResult = &x
				msg, err := schema.ToJSON(schema.FromGo(r))
				if err != nil {
					return err
				}

				_ = msg
				l.broadcast.BroadcastToSession(x.ID, msg)
				return nil
			},
		}, projection.WithName("8. Publish to websocket"))

	interpretation := projection.DefaultInMemoryInterpreter()
	return interpretation.Run(ctx, dag.Build())
}

func (l *LiveSelect) fromTyped(record schemaless.Record[tictactoemanage.State]) *schema.Map {
	return schema.MkMap(
		schema.MkField("ID", schema.MkString(record.ID)),
		schema.MkField("Type", schema.MkString(record.Type)),
		schema.MkField("Data", schema.FromGo(record.Data)),
		schema.MkField("Version", schema.MkInt(int(record.Version))),
	)
}

func (l *LiveSelect) fromUnTyped(record schemaless.Record[schema.Schema]) *schema.Map {
	return schema.MkMap(
		schema.MkField("ID", schema.MkString(record.ID)),
		schema.MkField("Type", schema.MkString(record.Type)),
		schema.MkField("Data", record.Data),
		schema.MkField("Version", schema.MkInt(int(record.Version))),
	)
}
