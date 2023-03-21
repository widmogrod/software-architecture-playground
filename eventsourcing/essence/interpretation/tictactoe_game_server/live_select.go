package tictactoe_game_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/projection"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

//go:generate moq -out live_select_broadcaster_moq_test.go . Broadcaster

type Broadcaster interface {
	BroadcastToSession(sessionID string, msg []byte)
}

type Stream interface {
	Subscribe(ctx context.Context, fromOffset int, f func(change schemaless.Change[schema.Schema])) error
}

type LiveSelect struct {
	stream    Stream
	store     schemaless.Repository[tictactoemanage.State]
	broadcast Broadcaster
	//root           *projection.DAGBuilder
	//streamState    projection.Builder
	//interpretation *projection.InMemoryInterpreter
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
		//interpretation: projection.DefaultInMemoryInterpreter(),
	}
}

func (l *LiveSelect) Process(ctx context.Context, sessionID string) error {
	//if l.root == nil {
	//	l.root = projection.NewDAGBuilder()
	//
	//	// Register streaming consumption only once
	//	l.streamState = l.root.
	//		//WithName("DynamoDB Filtered Stream").
	//		Load(&projection.GenerateHandler{
	//			Load: func(push func(message projection.Item)) error {
	//				log.Debugln("Load function called")
	//				return l.stream.Subscribe(ctx, 0, func(change schemaless.Change[schema.Schema]) {
	//					if change.Deleted {
	//						log.Warnf("Item was deleted: %v, live select skip on it", change)
	//						return
	//					}
	//
	//					record := *change.After
	//
	//					push(projection.Item{
	//						Key:  record.ID,
	//						Data: l.fromUnTyped(record),
	//					})
	//				})
	//			},
	//		}, projection.WithName("DynamoDB Stream"))
	//}

	where := predicate.MustWhere(
		"Data.SessionInGame.SessionID = :sessionID AND Type = :type",
		map[predicate.BindValue]schema.Schema{
			":sessionID": schema.FromGo(sessionID),
			":type":      schema.FromGo("game"),
		},
	)

	//dag := l.root
	dag := projection.NewDAGBuilder()
	// Only latest records from database that match live select criteria are used
	lastState := dag.
		//WithName("DynamoDB LastState Filtered").
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
		}, projection.WithName("Load DynamoDB LastState Filtered"))
	// Only streamed records that match live select criteria are used
	streamState := dag.
		Load(&projection.GenerateHandler{
			Load: func(push func(message projection.Item)) error {
				return l.stream.Subscribe(ctx, 0, func(change schemaless.Change[schema.Schema]) {
					if change.Deleted {
						log.Warnf("Item was deleted: %v, live select skip on it", change)
						return
					}

					record := *change.After

					push(projection.Item{
						Key:  record.ID,
						Data: l.fromUnTyped(record),
					})
				})
			},
		}, projection.WithName("DynamoDB Stream"))
	// Joining make sure that newest version is published

	//streamState, err := dag.GetByName("DynamoDB Filtered Stream")
	//if err != nil {
	//	log.Errorln("GetByName(DynamoDB Filtered Stream)", err)
	//	return err
	//}
	//streamState := l.streamState

	joined := dag.
		//WithName("Join DB & Stream").
		// Join by key, so if db and stream has the same key, then it will be joined.
		Join(lastState, streamState, projection.WithName("Join [DB & Stream]")).
		Map(&projection.FilterHandler{
			Where: where,
		}, projection.WithName("Filter[Join [DB & Stream]]")).
		// Joining by key and producing a new key is like merging!
		Merge(&projection.JoinHandler[schemaless.Record[tictactoemanage.State]]{
			F: func(a, b schemaless.Record[tictactoemanage.State], returning func(schemaless.Record[tictactoemanage.State])) error {
				if a.Version < b.Version {
					returning(b)
				}
				return nil
			},
		}, projection.WithName("Merge (version)[Join [DB & Stream]]"))

	gameStats := joined.
		//WithName("MapGameToStats").
		Map(&projection.MapHandler[schemaless.Record[tictactoemanage.State], tictactoemanage.SessionStatsResult]{
			F: func(x schemaless.Record[tictactoemanage.State], returning func(key string, value tictactoemanage.SessionStatsResult)) error {
				returning(GroupByKey(x.Data))
				return nil
			},
		}, projection.WithName("Map GameToStats"))

	gameStatsBySession := gameStats.
		//WithName("MergeSessionStats").
		Merge(&projection.MergeHandler[tictactoemanage.SessionStatsResult]{
			Combine: CombineByKey,
			DoRetract: func(base tictactoemanage.SessionStatsResult, x tictactoemanage.SessionStatsResult) (tictactoemanage.SessionStatsResult, error) {
				panic("retract not implemented")
			},
		}, projection.WithName("Merge SessionStats"))

	gameStatsBySession.
		//WithName("Publish to websocket, only when changed").
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
		}, projection.WithName("Publish to websocket"))

	interpretation := projection.DefaultInMemoryInterpreter()
	//interpretation := l.interpretation
	err := interpretation.Run(ctx, dag.Build())
	if err != nil {
		return err
	}
	//TODO figure out how to do closing down live select!
	//when connecion is closed
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
