package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/projection"
	//_ "net/http/pprof"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
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

func main() {
	flag.Parse()

	//// profile go program
	//f, err := os.Create("projection.cpu.prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if err := pprof.StartCPUProfile(f); err != nil {
	//	log.Fatal("could not start CPU profile: ", err)
	//}
	//defer pprof.StopCPUProfile()

	dag := projection.NewDAGBuilder()
	lines := dag.Load(&projection.GenerateHandler{
		func(push func(message projection.Item)) error {
			// open file and read lines and then push line
			file, err := os.Open(*input)
			if err != nil {
				return fmt.Errorf("failed to open file: %s %w", *input, err)
			}
			defer file.Close()

			// read lines and push
			fileScanner := bufio.NewScanner(file)
			fileScanner.Split(bufio.ScanLines)

			line := 0
			for fileScanner.Scan() {
				push(projection.Item{
					Key: *input,
					//Key:  fmt.Sprintf("line-%d", line),
					Data: schema.MkString(fileScanner.Text()),
				})
				line++
			}

			return nil
		},
	})

	words := lines.Map(&ExtractWordsFromLineHandler{
		SmallWordLength: *smallWordLength,
	})

	counted := words.Merge(&projection.MergeHandler[int]{
		Combine: func(a, b int) (int, error) {
			return a + b, nil
		},
	})

	formatted := counted.Map(&MapHandler{
		F: func(x projection.Item, returning func(value projection.Item)) error {
			returning(projection.Item{
				Key:  x.Key,
				Data: schema.MkString(formatFn(x.Key, schema.AsDefault[int](x.Data, 0))),
			})
			return nil
		},
	})

	var bufferedLines map[string]string = make(map[string]string)

	formatted.Map(&MapHandler{
		F: func(x projection.Item, returning func(value projection.Item)) error {
			line, ok := schema.As[string](x.Data)
			if !ok {
				return fmt.Errorf("failed to cast data to string. %#v", x.Data)
			}
			bufferedLines[x.Key] = line
			return nil
		},
	})

	interpreter := projection.DefaultInMemoryInterpreter()
	err := interpreter.Run(context.Background(), dag.Build())
	if err != nil {
		log.Fatalf("failed to run interpreter: %s", err)
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

// formatFn is a functional DoFn that formats a word and its count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

var _ projection.Handler = &ExtractWordsFromLineHandler{}

type ExtractWordsFromLineHandler struct {
	SmallWordLength int `json:"smallWordLength"`
	empty           atomic.Int32
	smallWords      atomic.Int32
}

func (e *ExtractWordsFromLineHandler) Process(x projection.Item, returning func(projection.Item)) error {
	line, ok := schema.As[string](x.Data)
	if !ok {
		return fmt.Errorf("failed to cast data to string")
	}

	//lineLen.Update(ctx, int64(len(line)))
	if len(strings.TrimSpace(line)) == 0 {
		e.empty.Add(1)
	}

	for _, word := range wordRE.FindAllString(line, -1) {
		// increment the counter for small words if length of words is
		// less than small_word_length
		if len(word) < e.SmallWordLength {
			e.smallWords.Add(1)
		}
		returning(projection.Item{
			Key:  word,
			Data: schema.MkInt(1),
		})
	}

	return nil
}

func (e *ExtractWordsFromLineHandler) Retract(x projection.Item, returning func(projection.Item)) error {
	//TODO implement me
	panic("implement me")
}

var _ projection.Handler = &MapHandler{}

type MapHandler struct {
	F func(x projection.Item, returning func(value projection.Item)) error
}

func (m *MapHandler) Process(x projection.Item, returning func(projection.Item)) error {
	return m.F(x, returning)
}

func (m *MapHandler) Retract(x projection.Item, returning func(projection.Item)) error {
	//TODO implement me
	panic("implement me")
}
