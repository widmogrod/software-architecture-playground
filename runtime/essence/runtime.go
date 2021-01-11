package main

import (
	"bytes"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation/eventsourcing"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	sdk := NewRuntimeSDK("http://localhost:8080")
	description, err := sdk.Describe()
	panicError(err)

	storage := eventsourcing.NewEventStore()

	c := cron.New()

	for _, schedule := range description.Schedule {
		schedule := schedule
		c.AddFunc(schedule.CRONInterval, func() {
			result := sdk.ScheduleInvoke(schedule, &runtime.ScheduleInvokeCMD{})
			log.Printf("[schedgule=%s] %s \n", schedule.HTTPEntrypoint, result.Logs)
		})
	}

	c.Start()

	mux := http.NewServeMux()

	var snapshot runtime.Snapshot

	for _, aggregate := range description.Aggregate {
		aggregate := aggregate
		mux.HandleFunc(aggregate.HTTPEntrypoint, func(w http.ResponseWriter, rq *http.Request) {
			//changes := make([]*runtime.AggregateChange, 0)
			//_ = storage.Reduce(func(change interface{}, result *eventsourcing.Reduced) *eventsourcing.Reduced {
			//	changes = append(changes, change.(*runtime.AggregateChange))
			//	return result
			//}, eventsourcing.Reduced{})
			//
			//// Todo applicator and reducer must be a pair!
			//reducer := &runtime.AggregateReducerType{
			//	AggregateType:  aggregate.AggregateType,
			//	HTTPEntrypoint: "/order/reduce",
			//}
			//output := sdk.AggregateReduce(reducer, &runtime.AggregateReduceCMD{
			//	AggregateRef: runtime.AggregateRef{
			//		ID:   rq.Header.Get("AggID"),
			//		Type: reducer.AggregateType,
			//	},
			//	Snapshot: nil,
			//	Changes:  changes,
			//})
			//
			payload, _ := ioutil.ReadAll(rq.Body)
			//if err != nil && err != io.EOF {
			//	output.Err = err.Error()
			//	output.Logs += "failed reading request body, err=" + err.Error()
			//	return
			//}

			input2 := &runtime.AggregateChangeCMD{
				AggregateID:   rq.Header.Get("AggID"),
				AggregateType: aggregate.AggregateType,
				Payload:       payload,
				Snapshot:      snapshot,
			}

			output2 := sdk.AggregateChange(aggregate, input2)
			if output2.Err == "" {
				snapshot = output2.Snapshot
			}

			defer json.NewEncoder(w).Encode(output2)

			// do storage
			for _, ch := range output2.Changes {
				log.Printf("record change(%#v)\n", ch)
				storage.Append(ch)
			}

			// output result
			//w.Write(output.Snapshot)
		})
	}

	//for _, reducer := range description.AggregateReducer {
	//	// TODO make it automatic, background not a API request
	//	// since this is a runtime responsibility
	//	reducer := reducer
	//	mux.HandleFunc(reducer.HTTPEntrypoint, func(w http.ResponseWriter, rq *http.Request) {
	//		changes := make([]*runtime.AggregateChange, 0)
	//		_ = storage.Reduce(func(change interface{}, result *eventsourcing.Reduced) *eventsourcing.Reduced {
	//			changes = append(changes, change.(*runtime.AggregateChange))
	//			return result
	//		}, eventsourcing.Reduced{})
	//
	//		output := sdk.AggregateReduce(reducer, &runtime.AggregateReduceCMD{
	//			AggregateRef: runtime.AggregateRef{
	//				ID:   rq.Header.Get("AggID"),
	//				Type: reducer.AggregateType,
	//			},
	//			Snapshot: nil,
	//			Changes:  changes,
	//		})
	//
	//		json.NewEncoder(w).Encode(output)
	//	})
	//}

	http.ListenAndServe(":8081", mux)
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

func NewRuntimeSDK(address string) *RuntimeSDK {
	return &RuntimeSDK{
		address: address,
		client: &http.Client{
			Timeout: 0,
		},
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

func (s *RuntimeSDK) ScheduleInvoke(schedule *runtime.ScheduleType, input *runtime.ScheduleInvokeCMD) *runtime.ScheduleInvokeResult {
	output := &runtime.ScheduleInvokeResult{}
	url := s.address + schedule.HTTPEntrypoint

	err := s.doRequest(url, input, output)
	if err != nil {
		output.Err = err.Error()
	}

	return output
}

func (s *RuntimeSDK) AggregateReduce(reducer *runtime.AggregateReducerType, input *runtime.AggregateReduceCMD) *runtime.AggregateReduceResult {
	output := &runtime.AggregateReduceResult{}
	url := s.address + reducer.HTTPEntrypoint

	err := s.doRequest(url, input, output)
	if err != nil {
		output.Err = err.Error()
	}

	return output
}

func (s *RuntimeSDK) AggregateChange(reducer *runtime.AggregateChangeType, input *runtime.AggregateChangeCMD) *runtime.AggregateChangeResult {
	output := &runtime.AggregateChangeResult{}
	url := s.address + reducer.HTTPEntrypoint

	err := s.doRequest(url, input, output)
	if err != nil {
		output.Err = err.Error()
	}

	return output
}

func (s *RuntimeSDK) doRequest(url string, input, output interface{}) error {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(input)
	if err != nil {
		return err
	}

	result, err := s.client.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	err = json.NewDecoder(result.Body).Decode(&output)
	if err != nil {
		return err
	}

	return nil
}
