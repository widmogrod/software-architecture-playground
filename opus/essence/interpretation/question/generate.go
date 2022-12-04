package question

//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/golang/protobuf/protobuf --go-grpc_out=. --go_out=. question.proto
