package tictactoe_game_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"io"
	"net/http"
	"net/url"
	"time"
)

type LiveSelectRequest struct {
	SessionID string
}

func NewLiveSelectClient(endpoint string) (*LiveSelectClient, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &LiveSelectClient{
		endpoint: endpointURL,
		client:   &http.Client{},
	}, nil
}

type LiveSelectClient struct {
	endpoint *url.URL
	client   *http.Client
}

func (l *LiveSelectClient) Process(ctx context.Context, sessionID string) error {
	body, err := json.Marshal(LiveSelectRequest{
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("liveselect.Process: json encoding %w", err)
	}

	url := *l.endpoint
	url.Path = "/live-select-process"

	response, err := l.client.Do(&http.Request{
		Method: http.MethodPost,
		URL:    &url,
		Body:   io.NopCloser(bytes.NewBuffer(body)),
	})
	if err != nil {
		return fmt.Errorf("liveselect.Process: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("liveselect.Process: status code not OK")
	}

	return nil
}

func (l *LiveSelectClient) Push(ctx context.Context, data []byte) error {
	url := *l.endpoint
	url.Path = "/live-select-push"

	response, err := l.client.Do(&http.Request{
		Method: http.MethodPost,
		URL:    &url,
		Body:   io.NopCloser(bytes.NewBuffer(data)),
	})
	if err != nil {
		return fmt.Errorf("liveselect.Push: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("liveselect.Push: status code not OK")
	}

	return nil
}

func NewLiveSelectServer(liveSelect *LiveSelect) *LiveSelectServer {
	return &LiveSelectServer{
		liveSelect:    liveSelect,
		maxSelectTime: 5 * time.Minute,
		workerQueue:   make(chan LiveSelectRequest),
	}
}

type LiveSelectServer struct {
	workerQueue   chan LiveSelectRequest
	liveSelect    *LiveSelect
	maxSelectTime time.Duration
}

func (server *LiveSelectServer) ProcessServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Info("🌀live-select: REQUEST")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "🌀live-select:error(1): %s", err)
		log.Errorln("🌀live-select:error(1): ", err)
		return
	}

	schemed, err := schema.FromJSON(body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "🌀live-select:error(2): %s", err)
		log.Errorln("🌀live-select:error(2): ", err)
		return
	}

	re, err := schema.ToGoG[LiveSelectRequest](schemed)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "🌀live-select:error(3): %s", err)
		log.Errorln("🌀live-select:error(3): ", err)
		return
	}

	server.workerQueue <- re

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(writer, "ok")
}

func (server *LiveSelectServer) Start(ctx context.Context) {
	log.Infof("🌀 live-select: BACKGROUND")
	defer log.Infof("🌀 live-select: BACKGROUND END")
	server.workerQueue = make(chan LiveSelectRequest)
	// TODO this will not trigger, when below for loop is running
	defer close(server.workerQueue)

	for req := range server.workerQueue {
		log.Infof("🌀 live-select: JOB : %v", req)
		go func(req LiveSelectRequest) {
			log.Infof("🌀 live-select: PROCESSING: %v", req)

			ctx2, _ := context.WithTimeout(ctx, server.maxSelectTime)
			err := server.liveSelect.Process(ctx2, req.SessionID)
			if err != nil {
				log.Errorf("🌀 live-select: ERR %s ", err)
			}
		}(req)
	}
}

func (server *LiveSelectServer) DynamoDBStreamServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "🧜‍DynamoDBStreamServeHTTP: error(1): %s", err)
		log.Errorln("🧜‍DynamoDBStreamServeHTTP: error(1): ", err)
		return
	}

	log.Info("🧜‍DynamoDBStreamServeHTTP: REQUEST", string(body))

	schemed, err := schema.FromJSON(body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "🧜‍DynamoDBStreamServeHTTP: error(2): %s", err)
		log.Errorln("🧜‍DynamoDBStreamServeHTTP: error(2): ", err)
		return
	}

	log.Info("🧜‍DynamoDBStreamServeHTTP: SCHEMED", schema.Get(schemed, "Records"))
	for _, record := range schema.Get(schemed, "Records").(*schema.List).Items {

		// potentially change, can be just state. And data pipeline can detect it
		// groub by  key
		// no initial state, not created
		// ther is state, and there is new - updated
		// there is state, with deleted flag set to true, delete
		// this implice that soft delete, or other options can happen.
		// but when is deleted, key could be closed? this would require some instruction
		// or maybe as with imposibility of distributed consensus,
		// data flush is important, windowing and triggers, etc?
		result := schemaless.Change[schema.Schema]{
			Before:  nil,
			After:   nil,
			Deleted: false,
		}

		switch schema.AsDefault[string](schema.Get(record, "eventName"), "") {
		case "MODIFY":
			// has both NewImage and OldImage
			old := schema.Get(record, "dynamodb.OldImage")
			before, err := server.toTyped(old)
			if err != nil {
				panic(err)
			}
			result.Before = &before

			new := schema.Get(record, "dynamodb.NewImage")
			after, err := server.toTyped(new)
			if err != nil {
				panic(err)
			}
			result.After = &after

		case "INSERT":
			// has only NewImage
			new := schema.Get(record, "dynamodb.NewImage")
			after, err := server.toTyped(new)
			if err != nil {
				panic(err)
			}
			result.After = &after
		case "REMOVE":
			// has only OldImage
			old := schema.Get(record, "dynamodb.OldImage")
			before, err := server.toTyped(old)
			if err != nil {
				panic(err)
			}
			result.Before = &before
			result.Deleted = true

		default:
			log.Errorln("🧜‍DynamoDBStreamServeHTTP: error(3): ",
				fmt.Errorf("unknown event name: %s",
					schema.AsDefault[string](schema.Get(schemed, "eventName"), "")))
			continue
		}

		err = server.liveSelect.Push(context.Background(), result)
		if err != nil {
			log.Errorln("🧜‍DynamoDBStreamServeHTTP: error(4): ", err)
		}
	}

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(writer, "🧜‍DynamoDBStreamServeHTTP: OK")
	log.Infoln("🧜‍DynamoDBStreamServeHTTP: OK")
}

func (server *LiveSelectServer) toTyped(record schema.Schema) (schemaless.Record[schema.Schema], error) {
	normalised, err := schema.UnwrapDynamoDB(record)
	if err != nil {
		data, err := schema.ToJSON(record)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return schemaless.Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream unwrap DynamoDB record: %v", record)
	}

	typed := schemaless.Record[schema.Schema]{
		ID:      schema.AsDefault[string](schema.Get(normalised, "ID"), "record-id-corrupted"),
		Type:    schema.AsDefault[string](schema.Get(normalised, "Type"), "record-id-corrupted"),
		Data:    schema.Get(normalised, "Data"),
		Version: schema.AsDefault[uint16](schema.Get(normalised, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		data, err := schema.ToJSON(normalised)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return schemaless.Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream corrupted record: %v", normalised)
	}
	return typed, nil
}
