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

	wshandler, err := tictactoe_game_server.NewWebSocket(context.Background())
	if err != nil {
		panic(err)
	}

	go wshandler.PublishLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	mux.HandleFunc("/play/", wshandler.ServeHTTP)

	handler := cors.AllowAll().Handler(mux)
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
