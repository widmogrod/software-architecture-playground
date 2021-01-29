package httpruntime

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
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
	Result
}

// ScheduleType type has all information necessary for runtime and client
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

// AggregateChangeType  type tha has all information necessary for runtime and client
// to define work on a aggregate
type AggregateChangeType struct {
	AggregateType  string
	CommandType    string
	HTTPEntrypoint string
}

type AggregateRef struct {
	ID   string
	Type string
}

type AggregateReduceCMD struct {
	AggregateRef AggregateRef

	Snapshot []byte
	Changes  []*AggregateChange
}

type Result struct {
	// Logs is a placeholder that collect logs during collected in a API invocation
	Logs string `json:"logs,omitempty"`
	Err  string `json:"err,omitempty"`
}

//TODO change back to transparent type
//Snapshot []byte
type Snapshot = json.RawMessage

type AggregateReduceResult struct {
	Result
	Snapshot Snapshot
}

type AggregateChange struct {
	Type       string
	Payload    []byte
	Version    uint
	RecordTime *time.Time
}

type AggregateChangeCMD struct {
	AggregateID   string
	AggregateType string

	Payload         []byte
	Snapshot        Snapshot
	SnapshotVersion uint
}

type AggregateChangeResult struct {
	Result

	AggregateType string
	AggregateID   string

	Changes  []*AggregateChange
	Snapshot Snapshot
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
	Schedule         []*ScheduleType
	Aggregate        []*AggregateChangeType
	AggregateReducer []*AggregateReducerType
}

func (r *MuxRuntimeClient) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	r.init.Do(func() {
		for _, b := range r.builders {
			r.mux.HandleFunc(
				b.Pattern,
				b.Handler,
			)

			// Just a convention with `continue`
			// to indicate that those types are unions
			// on when one is detected, then you don't check other

			if b.TypeSchedule != nil {
				r.entrypoint.Schedule = append(r.entrypoint.Schedule, b.TypeSchedule)
				continue
			}

			if b.TypeAggregateChange != nil {
				r.entrypoint.Aggregate = append(r.entrypoint.Aggregate, b.TypeAggregateChange)
				continue
			}

			if b.TypeAggregateReducer != nil {
				r.entrypoint.AggregateReducer = append(r.entrypoint.AggregateReducer, b.TypeAggregateReducer)
				continue
			}
		}

		r.mux.HandleFunc("/", func(w http.ResponseWriter, rq *http.Request) {
			json.NewEncoder(w).Encode(r.entrypoint)
		})
	})

	r.mux.ServeHTTP(w, rq)
}

type RequestTypeBuilder struct {
	Pattern string
	Handler func(w http.ResponseWriter, rq *http.Request)

	// union types below:
	TypeSchedule         *ScheduleType
	TypeAggregateChange  *AggregateChangeType
	TypeAggregateReducer *AggregateReducerType
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

func (b *RequestTypeBuilder) AggregateChange(aggregateType, commandType string) {
	b.TypeAggregateChange = &AggregateChangeType{
		AggregateType:  aggregateType,
		CommandType:    commandType,
		HTTPEntrypoint: b.Pattern,
	}
}

type AggregateReducerType struct {
	AggregateType  string
	HTTPEntrypoint string
}

func (b *RequestTypeBuilder) AggregateReducer(aggregateType string) {
	b.TypeAggregateReducer = &AggregateReducerType{
		AggregateType:  aggregateType,
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
