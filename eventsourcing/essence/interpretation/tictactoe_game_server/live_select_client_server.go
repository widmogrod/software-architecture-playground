package tictactoe_game_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
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

	response, err := l.client.Do(&http.Request{
		Method: http.MethodPost,
		URL:    l.endpoint,
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

func (server LiveSelectServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Info("live-select: REQUEST")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "live-select:error(1): %s", err)
		return
	}

	schemed, err := schema.FromJSON(body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "live-select:error(2): %s", err)
		return
	}

	re, err := schema.ToGoG[LiveSelectRequest](schemed)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "live-select:error(3): %s", err)
		return
	}

	server.workerQueue <- re

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(writer, "ok")
	log.Info("live-select: ", re)
}

func (server *LiveSelectServer) Start(ctx context.Context) {
	server.workerQueue = make(chan LiveSelectRequest)
	// TODO this will not trigger, when below for loop is running
	defer close(server.workerQueue)

	for req := range server.workerQueue {
		go func(req LiveSelectRequest) {
			log.Infof("livve-select: PROCESSING: %v", req)

			ctx2, _ := context.WithTimeout(ctx, server.maxSelectTime)
			err := server.liveSelect.Process(ctx2, req.SessionID)
			if err != nil {
				log.Errorf("%s ", err)
			}
		}(req)
	}
}
