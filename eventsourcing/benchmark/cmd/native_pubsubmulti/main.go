package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/projection"
	"runtime/pprof"
	"sync"

	//_ "net/http/pprof"
	"os"
	"regexp"
)

var (
	// By default, this example reads from a public dataset containing the text of
	// King Lear. Set this option to choose a different input file or glob.
	input = flag.String("input", "gs://apache-beam-samples/shakespeare/kinglear.txt", "File(s) to read.")

	// Set this required option to specify where to write the output.
	output = flag.String("output", "", "Output file (required).")
)

var (
	wordRE = regexp.MustCompile(`[a-zA-Z]+('[a-z])?`)
	//empty           = beam.NewCounter("extract", "emptyLines")
	smallWordLength = flag.Int("small_word_length", 9, "length of small words (default: 9)")
	//smallWords      = beam.NewCounter("extract", "smallWords")
	//lineLen         = beam.NewDistribution("extract", "lineLenDistro")
)

type Wc struct {
	Word  string
	Count int
}

func (wc *Wc) ToSchema() schema.Schema {
	return schema.MkMap(
		schema.MkField("Word", schema.MkString(wc.Word)),
		schema.MkField("Count", schema.MkInt(wc.Count)),
	)
}

func (wc *Wc) FromSchema(key string, value any) error {
	switch key {
	case "Word":
		wc.Word = any(value).(string)
	case "Count":
		wc.Count = int(any(value).(float64))
	default:
		return fmt.Errorf("unknown key %q", key)
	}
	return nil
}

func main() {
	flag.Parse()

	// profile go program
	f, err := os.Create("native_pubsubmulti.cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	f, err = os.Create("native_pubsubmulti.mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	defer f.Close()

	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	file, err := os.Open(*input)
	if err != nil {
		log.Fatalf("failed to open file: %s %v", *input, err)
	}
	defer file.Close()

	// read lines and push
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	var bufferedWords map[string]int = make(map[string]int)
	var bufferedLines map[string]string = make(map[string]string)

	multi := projection.NewPubSubMultiChan[string]()

	err = multi.Register("linesChannelPS")
	if err != nil {
		panic(err)
	}
	err = multi.Register("wordsCountChannelPS")
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wglines := sync.WaitGroup{}

	//for i := 0; i < 1; i++ {
	wg.Add(1)
	wglines.Add(1)
	go func() {
		defer wglines.Done()
		defer wg.Done()
		err := multi.Subscribe(context.Background(), "linesChannelPS", 0, func(x projection.Message) error {
			line := schema.AsDefault[string](x.Aggregate.Data, "error")
			for _, word := range wordRE.FindAllString(line, -1) {
				//data, err := json.Marshal(Wc{word, 1})
				//if err != nil {
				//	panic(err)
				//}

				r := Wc{word, 1}

				err := multi.Publish(
					context.Background(),
					"wordsCountChannelPS",
					projection.Message{
						Offset: 0,
						Key:    word,
						Aggregate: &projection.Item{
							Key: word,
							//Data: schema.MkBinary(data),
							//Data: schema.FromGo(Wc{word, 1}),
							Data: r.ToSchema(),
						},
						Retract: nil,
					},
				)
				if err != nil {
					panic(err)
				}
			}

			return nil
		})

		if err != nil {
			panic(err)
		}
	}()
	//}

	//for i := 0; i < 1; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ru := schema.WithOnlyTheseRules(
			//schema.WhenPath(nil, schema.UseStruct(Wc{})),
			schema.WhenPath(nil, NewCustomBuilder(func() FromSchemer {
				return &Wc{}
			})),
		)

		lock := sync.Mutex{}
		err := multi.Subscribe(context.Background(), "wordsCountChannelPS", 0, func(x projection.Message) error {
			//result := &Wc{}
			//_ = json.Unmarshal(x.Aggregate.Data.(*schema.Binary).B, result)
			res, err := schema.ToGo(x.Aggregate.Data, ru)
			if err != nil {
				panic(err)
			}
			result := any(res).(*Wc)

			//result, err := schema.ToGoG[Wc](
			//	x.Aggregate.Data,
			//)
			//if err != nil {
			//	panic(err)
			//}ult, err := schema.ToGoG[Wc](
			//	x.Aggregate.Data,
			//)
			//if err != nil {
			//	panic(err)
			//}

			lock.Lock()
			bufferedWords[result.Word] += result.Count
			//bufferedWords[x.Aggregate.Key] += 1
			lock.Unlock()
			return nil
		})
		if err != nil {
			panic(err)
		}

		//wordsCountChannelPS.Subscribe(func(Wc Wc) error {
		//	lock.Lock()
		//	bufferedWords[Wc.Word] += Wc.Count
		//	lock.Unlock()
		//	return nil
		//})
	}()
	//}

	for fileScanner.Scan() {
		line := fileScanner.Text()
		err := multi.Publish(context.Background(), "linesChannelPS", projection.Message{
			Offset: 0,
			Key:    line,
			Aggregate: &projection.Item{
				Key:  line,
				Data: schema.MkString(line),
			},
		})
		if err != nil {
			panic(err)
		}
	}
	multi.Finish(context.Background(), "linesChannelPS")
	//multi.Close()

	wglines.Wait()
	multi.Finish(context.Background(), "wordsCountChannelPS")
	//wordsCountChannelPS.Close()

	wg.Wait()

	//lock2 := sync.Mutex{}
	for word, count := range bufferedWords {
		//lock2.Lock()
		bufferedLines[word] = formatFn(word, count)
		//lock2.Unlock()
	}

	outputFile, err := os.Create(*output)
	if err != nil {
		log.Fatalf("failed to create output file: %s", err)
	}
	defer outputFile.Close()

	buf := bufio.NewWriterSize(outputFile, 1<<20) // use 1MB buffer
	for _, line := range bufferedLines {
		if _, err := buf.WriteString(line); err != nil {
			log.Fatalf("failed to write line: %s", err)
		}
		if _, err := buf.Write([]byte{'\n'}); err != nil {
			log.Fatalf("failed to write newline: %s", err)
		}
	}

	if err := buf.Flush(); err != nil {
		log.Fatalf("failed to flush buffer: %s", err)
	}
}

// formatFn is a functional DoFn that formats a Word and its Count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

type CustomBuiler struct {
	new func() FromSchemer
}

type FromSchemer interface {
	FromSchema(key string, value any) error
}

func (c *CustomBuiler) NewMapBuilder() schema.MapBuilder {
	return &customMapBuilder{
		fs: c.new(),
	}
}

var _ schema.TypeMapDefinition = (*CustomBuiler)(nil)

func NewCustomBuilder(new func() FromSchemer) *CustomBuiler {
	return &CustomBuiler{
		new: new,
	}
}

var _ schema.MapBuilder = (*customMapBuilder)(nil)

type customMapBuilder struct {
	fs FromSchemer
}

func (c *customMapBuilder) Set(key string, value any) error {
	return c.fs.FromSchema(key, value)
}

func (c *customMapBuilder) Build() any {
	return c.fs
}
