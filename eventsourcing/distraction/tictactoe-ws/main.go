package main

import (
	"context"
	"github.com/rs/cors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
	"log"
	"net/http"
	"os"
)

func main() {
	os.Setenv("AWS_PROFILE", "gh-dev")
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")

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
