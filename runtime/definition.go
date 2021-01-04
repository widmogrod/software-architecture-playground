package runtime

import (
	"encoding/json"
	"net/http"
	"sync"
)

// ScheduleInvokeCMD is a command payload that is send to invoke handler
// information in this command are available in a user defined handler.
// When runtime is a HTTP server, then it means that this structure represent body payload
type ScheduleInvokeCMD struct {
}

// ScheduleInvokeResult is a result that a user define handler can return, this result may indicate
// whenever process complete successfully, need to be retry, or just terminated
// When runtime is a HTTP server, then this means that runtime will attempt to unmarshall response to this payload
// and when it is possible, take action
type ScheduleInvokeResult struct {
	// Logs is a placeholder that collect logs during collected in a API invocation
	Logs string
}

// ScheduleType runtime type that has all information necessary for runtime and client
// to define work on schedule work
type ScheduleType struct {
	// CRONInterval follows this specification
	// ┌───────────── minute (0 - 59)
	// │ ┌───────────── hour (0 - 23)
	// │ │ ┌───────────── day of the month (1 - 31)
	// │ │ │ ┌───────────── month (1 - 12)
	// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday;
	// │ │ │ │ │                                   7 is also Sunday on some systems)
	// │ │ │ │ │
	// │ │ │ │ │
	// * * * * * <command to execute>
	CRONInterval string
	// HTTPEntrypoint represent how to invoke a user defined handler
	HTTPEntrypoint string
	// RetryTimes by default schedule won't be retry on failure
	RetryTimes uint
}

// EventSourcingRequest represent request
type EventSourcingRequest struct {
	AggregateID string
	Changes     [][]byte
}

// EventSourcingResponse contains result user define command handler
// runtime will take care of persisting result, changes, etc
//
// Handler on operation will have change set delivered
type EventSourcingResponse struct {
	AggregateType string
	AggregateID   string
	// Error or not, it's a result
	Result  []byte
	Logs    []byte
	Changes [][]byte
}

type ComposeProjectionRequest struct {
	AggregateType string
	AggregateID   string
	Changes       []byte
}

type ComposeProjectionResponse struct {
	Name           string
	TimeWindow     string
	RequestChanges []struct {
		AggregateType string
		// State or Change defines whenever send last state of aggregate
		// or just changes to it
		State *struct {
		}
		Change *struct {
		}
	}
	HTTPEntrypoint string
}

type MuxRuntimeClient struct {
	mux        *http.ServeMux
	init       sync.Once
	builders   []*RequestTypeBuilder
	entrypoint *RuntimeDescription
}

type RuntimeDescription struct {
	Schedule []*ScheduleType
}

func (r *MuxRuntimeClient) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	r.init.Do(func() {
		for _, b := range r.builders {
			// Schedule right now, but will be more
			r.mux.HandleFunc(
				b.Pattern,
				b.Handler,
			)

			if b.TypeSchedule != nil {
				r.entrypoint.Schedule = append(r.entrypoint.Schedule, b.TypeSchedule)
			}
		}

		r.mux.HandleFunc("/", func(w http.ResponseWriter, rq *http.Request) {
			json.NewEncoder(w).Encode(r.entrypoint)
		})
	})

	r.mux.ServeHTTP(w, rq)
}

type RequestTypeBuilder struct {
	Pattern      string
	Handler      func(w http.ResponseWriter, rq *http.Request)
	TypeSchedule *ScheduleType
}

type RequestTypeSchedule struct {
	Interval string
}

func (b *RequestTypeBuilder) Schedule(interval string) {
	b.TypeSchedule = &ScheduleType{
		CRONInterval:   interval,
		HTTPEntrypoint: b.Pattern,
	}
}

func (r *MuxRuntimeClient) HandleFunc(pattern string, handler func(w http.ResponseWriter, rq *http.Request)) *RequestTypeBuilder {
	builder := &RequestTypeBuilder{
		Pattern: pattern,
		Handler: handler,
	}

	r.builders = append(r.builders, builder)

	return builder
}

func NewMuxRuntimeClient() *MuxRuntimeClient {
	return &MuxRuntimeClient{
		mux:        http.NewServeMux(),
		builders:   []*RequestTypeBuilder{},
		entrypoint: &RuntimeDescription{},
	}
}
