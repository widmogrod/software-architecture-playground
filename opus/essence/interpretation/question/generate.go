package question

//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/golang/protobuf/protobuf --go-grpc_out=. --go_out=. question.proto

// next example is about generating custom code from proto files
//go:generate go run ../cmd/mms/main.go
