package main

import (
	"context"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
	"net/http"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "",
		PadLevelText:    true,
	})

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunLocalAWS,
	)

	ctx := context.Background()

	liveSelect := di.GetLiveSelectServer()
	go liveSelect.Start(ctx)

	wshandler := di.GetGolangWebSocketGameServer()
	go wshandler.PublishLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	mux.HandleFunc("/play/", wshandler.ServeHTTP)
	mux.HandleFunc("/live-select-process", liveSelect.ProcessServeHTTP)
	mux.HandleFunc("/live-select-push", liveSelect.DynamoDBStreamServeHTTP)

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
