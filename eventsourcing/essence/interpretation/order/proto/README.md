Install
```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc
```
Generate
``` 
protoc --go_out=plugins=grpc:. *.proto
```