package main

import (
	"context"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/protoorder"
	"google.golang.org/grpc"
	"net"
)

func _() {
	conn, _ := grpc.Dial("http://")
	client := protoorder.NewOrderAggregateClient(conn)
	state, _ := client.CreateOrder(nil, &protoorder.CreateOrderRequest{
		OrderID:  "",
		UserID:   "",
		Quantity: "",
	})

	fmt.Printf("client.CreateOrder() = %v", state)
}

func __() {
	ln, _ := net.Listen("tcp", ":8080")
	server := grpc.NewServer()
	protoorder.RegisterOrderAggregateServer(server, &protoorder.UnimplementedOrderAggregateServer{})
	_ = server.Serve(ln)
}

var _ protoorder.OrderAggregateServer = &MyServer{}

type MyServer struct{}

func (m *MyServer) CreateOrder(ctx context.Context, request *protoorder.CreateOrderRequest) (*protoorder.OrderAggregateState, error) {
	panic("implement me")
}
