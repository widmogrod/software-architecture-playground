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
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     false,
		TimestampFormat: "",
	})

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunAWS,
	)

	ctx := context.Background()

	liveSelect := di.GetLiveSelectServer()
	go liveSelect.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, rq *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/live-select", liveSelect.ServeHTTP)
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		sessionID := request.URL.Query().Get("sessionID")
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(writer, "OK, send to %s", sessionID)
		go di.GetBroadcaster().BroadcastToSession(sessionID, []byte("Hello from test"))
	})

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":80", handler)
	if err != nil {
		log.Fatal(err)
	}
}
