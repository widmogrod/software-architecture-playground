#!/bin/sh
set -e

# switch between binaries to run different benchmarks
# avaliable binaries: beam, native_count, native_gorutines, projection
#INPUT="kinglear.txt"
INPUT="kinglear.small.txt"

# shellcheck disable=SC2112
function output() {
    echo "$INPUT.$1.out"
}

cmd=$1

case $cmd in
    beam)
        time ./bin/beam --input $INPUT --output $(output "$cmd")
        ;;
    native_count)
        time ./bin/native_count --input $INPUT --output $(output "$cmd")
        ;;
    native_gorutines)
        time ./bin/native_gorutines --input $INPUT --output $(output "$cmd")
        ;;
    native_gorutines_select)
        time ./bin/native_gorutines_select --input $INPUT --output $(output "$cmd")
        ;;
    native_gorutines_pubsub)
        time ./bin/native_gorutines_pubsub --input $INPUT --output $(output "$cmd")
        ;;
    native_pubsubmulti)
        time ./bin/native_pubsubmulti --input $INPUT --output $(output "$cmd")
        ;;
    projection)
        time ./bin/projection --input $INPUT --output $(output "$cmd")
        ;;
    projection_slim)
        time ./bin/projection_slim --input $INPUT --output $(output "$cmd")
        ;;
    *)
        echo "Available commands:
         beam
         native_count
         native_gorutines
         native_gorutines_select
         native_gorutines_pubsub
         native_pubsubmulti
         projection
         projection_slim
         "

        ;;
esac