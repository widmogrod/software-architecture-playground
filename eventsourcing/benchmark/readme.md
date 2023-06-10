# Benchmark word count example

## 
- PubSubChan speed up whole process, and make it so that on 2,5GB file, beam runner swaps like crazy, and just cannot finish, but projection completes
- Quite x/schema ser-de with reflection is quite big bottleneck, but when change to static definition, finish only 2-3x slower than native implementation, which is 30x improvement
- 
## Some notes on different runs

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

     with buffered channels 1000
    4737.46s user 748.16s system 394% cpu 23:09.31 total

-- native count

go build -o nativecount native_count.go
time ./nativecount --input ./kinglear.txt --output kinglear.txt.native.out

2,5GB file (2686219278)

    95.97s user 1.54s system 101% cpu 1:36.31 total

-- native gorutines count

go build -o native_gorutines native_gorutines.go
time ./native_gorutines  --input ./kinglear.txt --output kinglear.txt.gorutimes.out

     173.29s user 52.89s system 183% cpu 2:02.94 total


###
(software-architecture-playground) ➜  benchmark git:(feature/february-2) ✗ ./run.sh native_count

real    1m36.972s
user    1m36.566s
sys     0m1.814s
(software-architecture-playground) ➜  benchmark git:(feature/february-2) ✗ ./run.sh native_gorutines

real    2m2.860s
user    2m55.523s
sys     0m53.721s
(software-architecture-playground) ➜  benchmark git:(feature/february-2) ✗ ./run.sh native_gorutines_pubsub

real    2m29.568s
user    3m44.729s
sys     0m52.278s

(software-architecture-playground) ➜  benchmark git:(feature/february-2) ✗  ./run.sh projection

real    13m24.520s
user    41m22.943s
sys     8m43.810s
