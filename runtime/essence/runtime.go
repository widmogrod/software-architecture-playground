package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"net/http"
)

func main() {
	client := NewRuntimeSDK("http://localhost:8080")
	description, err := client.Describe()
	panicError(err)

	c := cron.New()

	for _, schedule := range description.Schedule {
		call := func(schedule *runtime.ScheduleType) func() {
			return func() {
				url := client.address + schedule.HTTPEntrypoint
				input := runtime.ScheduleInvokeCMD{}
				output := runtime.ScheduleInvokeResult{}

				body := &bytes.Buffer{}
				json.NewEncoder(body).Encode(input)
				result, err := client.client.Post(url, "application/json", body)
				if err != nil {
					panicError(err)
				}

				err = json.NewDecoder(result.Body).Decode(&output)
				if err != nil {
					panicError(err)
				}

				fmt.Println(output.Logs)
			}
		}

		c.AddFunc(schedule.CRONInterval, call(schedule))
	}

	c.Run()
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
