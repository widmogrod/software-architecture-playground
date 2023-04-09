# Benchmark word count example

go install github.com/apache/beam/sdks/v2/go/examples/wordcount

curl https://raw.githubusercontent.com/apache/beam/master/sdks/go/data/shakespeare/kinglear.txt > kinglear.txt

time $GOPATH/bin/wordcount  --input ./kinglear.txt --output kinglear.txt.beam.out

0.06s user 0.03s system 85% cpu 0.107 total
0.06s user 0.01s system 152% cpu 0.046 total

2,5GB file (2686219278)

    376.18s user 421.81s system 101% cpu 13:09.09 total

---

go build -o projectioncount projection.go
time ./projectioncount   --input ./kinglear.txt --output kinglear.txt.projection.out

0.17s user 0.03s system 213% cpu 0.094 total
0.23s user 0.05s system 233% cpu 0.122 total

2,5GB file (2686219278)

    not finished, with inmemory pubsub base on list

    channel based pubsub
     4952.98s user 758.25s system 25% cpu 6:14:12.08 total

-- native count

go build -o nativecount native_count.go
time ./nativecount --input ./kinglear.txt --output kinglear.txt.native.out

2,5GB file (2686219278)

    95.97s user 1.54s system 101% cpu 1:36.31 total