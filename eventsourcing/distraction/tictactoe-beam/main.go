package main

import (
	"context"
	"flag"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
	"log"
)

func main() {
	flag.Parse()
	beam.Init()

	p, root := beam.NewPipelineWithRoot()

	enchanced := beam.ParDo(root.Scope("Words"), func(x []byte, emit func(string)) {
		emit(string(x) + "word1")
		emit(string(x) + "word2")
	}, beam.Impulse(root))

	beam.ParDo0(root.Scope("Log"), func(x string) {
		log.Println("word:", x)
	}, enchanced)

	err := beamx.Run(context.Background(), p)
	if err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
