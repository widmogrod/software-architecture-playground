package main

import (
	"context"
	"fmt"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
	"net/http"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      false,
		DisableQuote:     true,
		DisableTimestamp: true,
	})

	defer func() {
		if r := recover(); r != nil {
			log.Errorln("LiveSelectServer panic", r)
		}
	}()
	defer log.Infoln("LiveSelectServer stopped gracefully")

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunAWS,
	)

	ctx := context.Background()

	liveSelect := di.GetLiveSelectServer()
	go liveSelect.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, rq *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/live-select-process", liveSelect.ProcessServeHTTP)
	mux.HandleFunc("/live-select-push", liveSelect.DynamoDBStreamServeHTTP)

	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		sessionID := request.URL.Query().Get("sessionID")
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(writer, "OK, send to %s", sessionID)
		go di.GetBroadcaster().BroadcastToSession(sessionID, []byte("Hello from test"))
	})

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
