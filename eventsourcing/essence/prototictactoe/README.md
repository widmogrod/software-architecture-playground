Install
```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc
```
Generate

For: Go
``` 
protoc --go_out=plugins=grpc:. *.proto
```

For: JS
```
npm install protoc-gen-grpc-web -g
protoc --js_out=import_style=commonjs:. \
    --grpc-web_out=import_style=typescript,mode=grpcwebtext:. \
    *.proto
```

```
npm install -g grpc-tools
npm install -g protoc-gen-ts
npm install -g grpc_tools_node_protoc_ts

grpc_tools_node_protoc --js_out=import_style=commonjs,binary:. --grpc_out=grpc_js:. *.proto

[comment]: <> (grpc_tools_node_protoc --plugin=protoc-gen-ts=`which protoc-gen-ts` --ts_out=grpc_js:. *.proto)

grpc_tools_node_protoc --js_out=import_style=commonjs,binary:. --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` *.proto
protoc --plugin=protoc-gen-ts=`which protoc-gen-ts` --ts_out=. *.proto

grpc_tools_node_protoc --js_out=import_style=commonjs,binary:. --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` *.proto
```


```
grpc_tools_node_protoc \
--js_out=import_style=commonjs,binary:./your_dest_dir \
--grpc_out=./your_dest_dir \
--plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` \
-I ./ \
proto.proto

protoc \
--plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts \
--ts_out=./your_dest_dir \
-I ./ \
proto.proto

```