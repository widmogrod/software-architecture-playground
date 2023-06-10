package main

import (
	eventstore "github.com/EventStore/EventStore-Client-Go/client"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate/store/eventstoredb"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/deserializer"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe/proto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoeaggregate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {
	ln, _ := net.Listen("tcp", ":8080")
	server := grpc.NewServer()
	reflection.Register(server)

	conf, err := eventstore.ParseConnectionString("esdb://admin:changeit@localhost:2113?tls=false&tlsverifycert=false")
	if err != nil {
		panic(err)
	}

	client, err := eventstore.NewClient(conf)
	if err != nil {
		panic(err)
	}

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	ser := deserializer.NewDeSerializer()
	ser.Register(tictactoeaggregate.GameFinish{})
	ser.Register(tictactoeaggregate.Moved{})
	ser.Register(tictactoeaggregate.SecondPlayerJoined{})
	ser.Register(tictactoeaggregate.GameCreated{})

	st := eventstoredb.NewEventStoreDB(client, ser)

	tictactoe := tictactoe.NewTicTacToeServer(aggregate.NewAggregate(func() aggregate.Aggregate {
		return tictactoeaggregate.NewTicTacToeAggregate()
	}, st))

	proto.RegisterTicTacToeAggregateServer(server, tictactoe)
	_ = server.Serve(ln)
}
