package main

import (
	"encoding/json"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"io"
	"net/http"
	"time"
)

func main() {
	// what if data is polymorfic
	// what if everything is an networks call
	// what if response contains data to be persisted, events to be publish, ...
	// what if storage is not a user defined concern?
	// what if external communication also is HTTP base, and persisting results must always happen in aggregates
	//  and external information can be only delivered as a webhook

	//mux := http.NewServeMux()
	//mux.HandleFunc("/entrypoint/schedule", func(rq http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(rq, "Hohohoho")
	//})
	app := runtime.NewMuxRuntimeClient()
	app.
		HandleFunc("/schedule/time", func(w http.ResponseWriter, rq *http.Request) {
			input := runtime.ScheduleInvokeCMD{}
			output := runtime.ScheduleInvokeResult{}

			err := json.NewDecoder(rq.Body).Decode(&input)
			if err != nil && err != io.EOF {
				output.Logs = err.Error()
			} else {
				output.Logs = time.Now().String()
			}

			json.NewEncoder(w).Encode(output)
		}).
		Schedule("* * * * *")

	http.ListenAndServe(":8080", app)
}
