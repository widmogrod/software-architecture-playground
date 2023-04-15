go build -o bin/beam  ./cmd/beam
go build -o bin/native_count  ./cmd/native_count
go build -o bin/native_gorutines ./cmd/native_gorutines
go build -o bin/native_gorutines_select ./cmd/native_gorutines_select
go build -o bin/native_gorutines_pubsub ./cmd/native_gorutines_pubsub
go build -o bin/native_pubsubmulti ./cmd/native_pubsubmulti
go build -o bin/projection ./cmd/projection
go build -o bin/projection_slim ./cmd/projection_slim