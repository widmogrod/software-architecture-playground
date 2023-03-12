package tictactoe_game_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/projection"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"time"
)

//go:generate moq -out live_select_broadcaster_moq_test.go . Broadcaster

type Broadcaster interface {
	BroadcastToSession(sessionID string, msg []byte)
}

type Stream interface {
	Subscribe(fromOffset int) (schemaless.Change[schema.Schema], int)
	Subscribe2() <-chan schemaless.Change[schema.Schema]
}

type LiveSelect struct {
	stream    Stream
	store     schemaless.Repository[tictactoemanage.State]
	broadcast Broadcaster
}

func NewLiveSelect(
	stream Stream,
	store schemaless.Repository[tictactoemanage.State],
	broadcast Broadcaster,
) *LiveSelect {
	return &LiveSelect{
		stream:    stream,
		store:     store,
		broadcast: broadcast,
	}
}

func (l *LiveSelect) Process(ctx context.Context, sessionID string) error {
	where := predicate.MustWhere(
		"Data.SessionInGame.SessionID = :sessionID AND Type = :type",
		map[predicate.BindValue]schema.Schema{
			":sessionID": schema.FromGo(sessionID),
			":type":      schema.FromGo("game"),
		},
	)

	dag := projection.NewBuilder()
	// Only latest records from database that match live select criteria are used
	lastState := dag.
		WithName("DynamoDB LastState Filtered").
		Load(&projection.GenerateHandler{
			Load: func(push func(message projection.Item)) error {
				results, err := l.store.FindingRecords(schemaless.FindingRecords[schemaless.Record[tictactoemanage.State]]{
					Where: where,
				})
				log.Debugln("results", results)

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
		})
	// Only streamed records that match live select criteria are used
	streamState := dag.
		WithName("DynamoDB Filtered Stream").
		Load(&projection.GenerateHandler{
			Load: func(push func(message projection.Item)) error {
				//for {
				//	select {
				//	case <-ctx.Done():
				//		return nil
				//	case item := <-l.stream.Subscribe2():
				//		if item.Deleted {
				//			log.Warnf("Item was deleted: %v, live select skip on it", item)
				//			continue
				//		}
				//
				//		// we are interested in the latest state
				//		record := *item.After
				//
				//		push(projection.Item{
				//			Key:  record.ID,
				//			Data: l.fromUnTyped(record),
				//		})
				//	}
				//}
				var lastOffset int = 0
				for {
					select {
					case <-ctx.Done():
						return nil

					default:
						item, offset := l.stream.Subscribe(lastOffset)
						if offset == lastOffset {
							<-time.After(10 * time.Millisecond)
							continue
						}

						lastOffset = offset

						if item.Deleted {
							log.Warnf("Item was deleted: %v, live select skip on it", item)
							continue
						}

						// we are interested in the latest state
						record := *item.After

						push(projection.Item{
							Key:  record.ID,
							Data: l.fromUnTyped(record),
						})
					}
				}

				return nil
			},
		})
	// Joining make sure that newest version is published

	joined := dag.
		WithName("Join DB & Stream").
		// Join by key, so if db and stream has the same key, then it will be joined.
		Join(lastState, streamState).
		Map(&projection.FilterHandler{
			Where: where,
		}).
		// Joining by key and producing a new key is like merging!
		Merge(&projection.JoinHandler[schemaless.Record[tictactoemanage.State]]{
			F: func(a, b schemaless.Record[tictactoemanage.State], returning func(schemaless.Record[tictactoemanage.State])) error {
				if a.Version < b.Version {
					returning(b)
				}
				return nil
			},
		})

	gameStats := joined.
		WithName("MapGameToStats").
		Map(&projection.MapHandler[schemaless.Record[tictactoemanage.State], tictactoemanage.SessionStatsResult]{
			F: func(x schemaless.Record[tictactoemanage.State], returning func(key string, value tictactoemanage.SessionStatsResult)) error {
				returning(GroupByKey(x.Data))
				return nil
			},
		})

	gameStatsBySession := gameStats.
		WithName("MergeSessionStats").
		Merge(&projection.MergeHandler[tictactoemanage.SessionStatsResult]{
			Combine: CombineByKey,
			DoRetract: func(base tictactoemanage.SessionStatsResult, x tictactoemanage.SessionStatsResult) (tictactoemanage.SessionStatsResult, error) {
				panic("retract not implemented")
			},
		})

	gameStatsBySession.
		WithName("Publish to websocket, only when changed").
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
		})

	interpretation := projection.DefaultInMemoryInterpreter()
	err := interpretation.Run(dag.Build())
	if err != nil {
		return err
	}
	//TODO figure out how to do closing down live select!
	//when connecion is closed
	interpretation.WaitForDone()
	return nil
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
