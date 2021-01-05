package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation/eventsourcing"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"io"
	"io/ioutil"
	"net/http"
)

func main() {
	client := NewRuntimeSDK("http://localhost:8080")
	description, err := client.Describe()
	panicError(err)

	storage := eventsourcing.NewEventStore()

	c := cron.New()

	for _, schedule := range description.Schedule {
		call := func(schedule *runtime.ScheduleType) func() {
			return func() {
				url := client.address + schedule.HTTPEntrypoint
				input := &runtime.ScheduleInvokeCMD{}
				output := &runtime.ScheduleInvokeResult{}

				body := &bytes.Buffer{}
				json.NewEncoder(body).Encode(input)
				result, err := client.client.Post(url, "application/json", body)
				if err != nil {
					panicError(err)
				}

				err = json.NewDecoder(result.Body).Decode(output)
				if err != nil {
					panicError(err)
				}

				fmt.Println(output.Logs)
			}
		}

		c.AddFunc(schedule.CRONInterval, call(schedule))
	}

	c.Start()

	mux := http.NewServeMux()

	for _, aggregate := range description.Aggregate {
		mux.HandleFunc(aggregate.HTTPEntrypoint, func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateChangeCMD{}
			output := &runtime.AggregateChangeResult{}

			defer json.NewEncoder(w).Encode(output)

			payload, err := ioutil.ReadAll(rq.Body)
			if err != nil && err != io.EOF {
				output.Err = err
				output.Logs += "failed reading request body, err=" + err.Error()
				return
			}

			url := client.address + aggregate.HTTPEntrypoint
			input.Payload = payload
			input.AggregateType = aggregate.AggregateType

			body, err := json.Marshal(input)
			if err != nil {
				output.Err = err
				output.Logs += "failed marshalling AggregateChangeCMD, err=" + err.Error()
				return
			}

			result, err := client.client.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				output.Err = err
				output.Logs += "failed forwarding request to " + url + ", err=" + err.Error()
				return
			}

			//body, err = ioutil.ReadAll(result.Body)
			//output.Logs += fmt.Sprintf("body(%s)\n", body)

			err = json.NewDecoder(result.Body).Decode(&output)
			if err != nil {
				output.Err = err
				output.Logs += "failed decoding response to AggregateChangeResult struct, err=" + err.Error()
				return
			}

			// do storage

			for _, ch := range output.Changes {
				output.Logs += fmt.Sprintf("record change(%#v)\n", ch)
				storage.Append(ch)
			}

			// output result
			//w.Write(output.Result)

		})
	}

	for _, reducer := range description.AggregateReducer {
		// TODO make it automatic, background not a API request
		// since this is a runtime responsibility
		mux.HandleFunc(reducer.HTTPEntrypoint, func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateReduceCMD{}
			output := &runtime.AggregateReduceResult{}

			defer json.NewEncoder(w).Encode(output)

			url := client.address + reducer.HTTPEntrypoint
			input.AggregateType = reducer.AggregateType
			init := eventsourcing.Reduced{
				Value: input,
			}

			err = storage.
				Reduce(func(change interface{}, result *eventsourcing.Reduced) *eventsourcing.Reduced {
					output.Logs += fmt.Sprintf("change(%#v)\n", change)
					//input := result.Value.(*runtime.AggregateReduceCMD)
					input.Changes = append(input.Changes, change.(*runtime.AggregateChange))
					return result
				}, init).Err
			if err != nil {
				output.Err = err
				output.Logs += "failed collecting changes for AggregateReduceCMD, err=" + err.Error()
				return
			}

			body, err := json.Marshal(input)
			if err != nil {
				output.Err = err
				output.Logs += "failed marshalling AggregateChangeCMD, err=" + err.Error()
				return
			}

			output.Logs += string(body)
			result, err := client.client.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				output.Err = err
				output.Logs += "failed forwarding request to " + url + ", err=" + err.Error()
				return
			}
			//output.Snapshot = result

			err = json.NewDecoder(result.Body).Decode(&output)
			if err != nil {
				output.Err = err
				output.Logs += "failed decoding response to AggregateReduceResult struct, err=" + err.Error()
				return
			}

			//body, _ = ioutil.ReadAll(result.Body)
			//output.Snapshot = body

			// Persist snapshot
			//storage.Snapshot()
		})
	}

	http.ListenAndServe(":8081", mux)
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

type RuntimeSDK struct {
	client      *http.Client
	address     string
	description *runtime.RuntimeDescription
}

func (s *RuntimeSDK) Describe() (*runtime.RuntimeDescription, error) {
	output := &runtime.RuntimeDescription{}

	request, err := s.client.Get(s.address)
	if err != nil {
		return output, err
	} else {
		err = json.NewDecoder(request.Body).Decode(output)
	}

	return output, err
}

func NewRuntimeSDK(address string) *RuntimeSDK {
	return &RuntimeSDK{
		address: address,
		client: &http.Client{
			Timeout: 0,
		},
	}
}
