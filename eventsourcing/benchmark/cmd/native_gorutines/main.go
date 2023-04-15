package main

import (
	"bufio"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
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

type wc struct {
	word  string
	count int
}

func main() {
	flag.Parse()
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

	var linesChannel = make(chan string, 100)
	var wordsCountChannel = make(chan wc, 100)

	wg := sync.WaitGroup{}
	wglines := sync.WaitGroup{}

	for i := 0; i < 1; i++ {
		wg.Add(1)
		wglines.Add(1)
		go func() {
			defer wglines.Done()
			defer wg.Done()
			for line := range linesChannel {
				for _, word := range wordRE.FindAllString(line, -1) {
					wordsCountChannel <- wc{word, 1}
				}
			}
		}()
	}

	//lock := sync.Mutex{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for wc := range wordsCountChannel {
				//lock.Lock()
				bufferedWords[wc.word] += wc.count
				//lock.Unlock()
			}
		}()
	}

	for fileScanner.Scan() {
		line := fileScanner.Text()
		linesChannel <- line
	}
	close(linesChannel)

	wglines.Wait()
	close(wordsCountChannel)

	wg.Wait()

	for word, count := range bufferedWords {
		bufferedLines[word] = formatFn(word, count)
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
