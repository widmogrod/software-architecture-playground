package tictactoe_game_server

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/typedful"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"net/url"
	"os"
	"sync"
)

type Development int

const (
	RunLocalInMemoryDevelopment = iota
	RunLocalAWS
	RunAWS
)

type KeyCache struct {
	data map[string]any
	lock sync.RWMutex
}

func (cache *KeyCache) Get(key string, do func() any) any {
	cache.lock.RLock()
	if value, ok := cache.data[key]; ok {
		cache.lock.RUnlock()
		return value
	}
	cache.lock.RUnlock()

	value := do()

	cache.lock.Lock()
	cache.data[key] = value
	cache.lock.Unlock()
	return value
}

func DefaultDI(mode Development) *DI {
	return &DI{
		mode: mode,
		keyCache: KeyCache{
			data: make(map[string]any),
		},
	}
}

type DI struct {
	mode     Development
	keyCache KeyCache
}

func (di *DI) MustAWSConfig() aws.Config {
	return di.keyCache.Get("aws-config", func() any {
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}
		return cfg
	}).(aws.Config)
}

func (di *DI) GetDynamoDBStore() *schemaless.DynamoDBRepository {
	return di.keyCache.Get("dynamodb-repo", func() any {
		cfg := di.MustAWSConfig()
		return schemaless.NewDynamoDBRepository(dynamodb.NewFromConfig(cfg), di.GetDynamoDBTableName())
	}).(*schemaless.DynamoDBRepository)
}

func (di *DI) GetDynamoDBTableName() string {
	return di.keyCache.Get("dynamodb-table-name", func() any {
		tableName := os.Getenv("TABLE_NAME")
		if tableName == "" {
			tableName = "test-repo-record"
		}
		return tableName
	}).(string)
}

func (di *DI) GetKinesisStream() *schemaless.KinesisStream {
	return di.keyCache.Get("kinesis-stream", func() any {
		cfg := di.MustAWSConfig()
		stream := schemaless.NewKinesisStream(kinesis.NewFromConfig(cfg), di.GetKinesisStreamName())
		go stream.Process()
		return stream
	}).(*schemaless.KinesisStream)
}

func (di *DI) GetKinesisStreamName() string {
	return di.keyCache.Get("kinesis-stream-name", func() any {
		streamName := os.Getenv("KINESIS_STREAM_NAME")
		if streamName == "" {
			streamName = "test-record-stram"
		}
		return streamName
	}).(string)
}

func (di *DI) GetAWSWebSocketPublisher() *websockproto.AWSPublisher {
	return di.keyCache.Get("aws-websocket-publisher", func() any {
		domainName := os.Getenv("DOMAIN_NAME")

		domainURL, err := url.Parse(domainName)
		if err == nil {
			domainName = domainURL.Host
		}

		stageName := os.Getenv("STAGE_NAME")

		return websockproto.NewPublisher(
			domainName,
			stageName,
			di.MustAWSConfig(),
		)
	}).(*websockproto.AWSPublisher)
}

func (di *DI) GetInMemoryWebSocketPublisher() *websockproto.InMemoryProtocol {
	return di.keyCache.Get("in-memory-websocket-publisher", func() any {
		return websockproto.NewInMemoryProtocol()
	}).(*websockproto.InMemoryProtocol)
}

func (di *DI) GetConnectionToSessionRepository() *typedful.TypedRepoWithAggregator[websockproto.ConnectionToSession, any] {
	return di.keyCache.Get("connection-to-session-repo", func() any {
		return typedful.NewTypedRepository[websockproto.ConnectionToSession](di.GetStore())
	}).(*typedful.TypedRepoWithAggregator[websockproto.ConnectionToSession, any])
}

func (di *DI) GetInMemoryStore() *schemaless.InMemoryRepository {
	return di.keyCache.Get("in-memory-repo", func() any {
		return schemaless.NewInMemoryRepository()
	}).(*schemaless.InMemoryRepository)
}

func (di *DI) GetStore() schemaless.Repository[schema.Schema] {
	return di.keyCache.Get("store", func() any {
		switch di.mode {
		case RunLocalInMemoryDevelopment:
			return di.GetInMemoryStore()
		case RunLocalAWS, RunAWS:
			return di.GetDynamoDBStore()

		}

		panic("not reachable")
	}).(schemaless.Repository[schema.Schema])
}

func (di *DI) GetBroadcaster() *websockproto.InMemoryBroadcaster {
	return di.keyCache.Get("broadcaster", func() any {
		switch di.mode {
		case RunLocalInMemoryDevelopment, RunLocalAWS:
			return websockproto.NewBroadcaster(di.GetInMemoryWebSocketPublisher(), di.GetConnectionToSessionRepository())

		case RunAWS:
			return websockproto.NewBroadcaster(di.GetAWSWebSocketPublisher(), di.GetConnectionToSessionRepository())
		}

		panic("not reachable")
	}).(*websockproto.InMemoryBroadcaster)
}

func (di *DI) GetTicTacToeManageStateRepository() *typedful.TypedRepoWithAggregator[tictactoemanage.State, any] {
	return di.keyCache.Get("tictactoemanage-state-repo", func() any {
		return typedful.NewTypedRepository[tictactoemanage.State](di.GetStore())
		//return typedful.NewTypedRepoWithAggregator[tictactoemanage.State, tictactoemanage.SessionStatsResult](
		//	store,
		//	func() schemaless.Aggregator[tictactoemanage.State, tictactoemanage.SessionStatsResult] {
		//		return NewTictactoeManageStateAggregate(store)
		//	},
		//)
	}).(*typedful.TypedRepoWithAggregator[tictactoemanage.State, any])
}

func (di *DI) GetLiveSelect() *LiveSelect {
	return di.keyCache.Get("live-select", func() any {
		switch di.mode {
		case RunLocalInMemoryDevelopment:
			return NewLiveSelect(
				di.GetTicTacToeManageStateRepository(),
				di.GetBroadcaster(),
			).UseStreamToPush(di.GetInMemoryStore().AppendLog())

		case RunLocalAWS:
			return NewLiveSelect(
				di.GetTicTacToeManageStateRepository(),
				di.GetBroadcaster(),
			).UseStreamToPush(di.GetKinesisStream())

		case RunAWS:
			return NewLiveSelect(
				di.GetTicTacToeManageStateRepository(),
				di.GetBroadcaster(),
			)
		}

		panic("not reachable")
	}).(*LiveSelect)
}

func (di *DI) GetLiveSelectServerEndpoint() string {
	return di.keyCache.Get("live-select-server-endpoint", func() any {
		endpoint := os.Getenv("LIVE_SELECT_SERVER_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:8080"
		}
		return endpoint
	}).(string)
}

func (di *DI) GetLiveSelectClient() *LiveSelectClient {
	return di.keyCache.Get("live-select-client", func() any {
		client, err := NewLiveSelectClient(
			di.GetLiveSelectServerEndpoint(),
		)
		if err != nil {
			panic(err)
		}
		return client
	}).(*LiveSelectClient)
}

func (di *DI) GetLiveSelectServer() *LiveSelectServer {
	return di.keyCache.Get("live-select-server", func() any {
		return NewLiveSelectServer(
			di.GetLiveSelect(),
		)
	}).(*LiveSelectServer)
}

func (di *DI) GetQueryUsingStorage() Query {
	return di.keyCache.Get("query-using-storage", func() any {
		return NewQueryUsingStorage(
			typedful.NewTypedRepository[tictactoemanage.SessionStatsResult](di.GetStore()),
		)
	}).(Query)
}

func (di *DI) GetQueryUsingOpenSearch() (*OpenSearchStorage, error) {
	panic("not implemented")
	//return NewQuery(
	//	"https://search-dynamodb-projection-vggyq7lvwooliwe65oddc5gyse.eu-west-1.es.amazonaws.com/",
	//	"lambda-index",
	//)
}

func (di *DI) GetGame() *Game {
	return di.keyCache.Get("game", func() any {
		return &Game{
			broadcast:           di.GetBroadcaster(),
			gameStateRepository: di.GetTicTacToeManageStateRepository(),
			query:               di.GetQueryUsingStorage(),
			//liveSelect:          di.GetLiveSelect(),
			liveSelect: di.GetLiveSelectClient(),
		}
	}).(*Game)
}

func (di *DI) GetGolangWebSocketGameServer() *websockproto.InMemoryProtocol {
	game := di.GetGame()
	return di.keyCache.Get("server", func() any {
		wshandler := di.GetInMemoryWebSocketPublisher()
		wshandler.OnMessage = game.OnMessage
		wshandler.OnConnect = game.OnConnect
		wshandler.OnDisconnect = game.OnDisconnect
		return wshandler
	}).(*websockproto.InMemoryProtocol)
}
