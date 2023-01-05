package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/cors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

func main() {
	registry := sync.Map{}
	peers := sync.Map{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	//mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	// generate uuid for game
	//	uuid, err := uuid.NewUUID()
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		w.Write([]byte(err.Error()))
	//	} else {
	//		w.WriteHeader(http.StatusSeeOther)
	//		w.Header().Add("Location", "/play/"+uuid.String())
	//	}
	//})
	mux.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/play/")
		fmt.Println("uuid", uuid)
		//uuid := "game-1"
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		go func() {
			defer conn.Close()

			connectionsAny, _ := peers.LoadOrStore(uuid, &sync.Map{})
			connections := connectionsAny.(*sync.Map)
			connections.Store(conn, conn)
			defer connections.Delete(conn)

			gameAny, _ := registry.LoadOrStore(uuid, tictacstatemachine.NewMachine())
			game := gameAny.(*tictacstatemachine.Machine)

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					if err == io.EOF {
						return
					}
					log.Printf("read error: %v", err)
					continue
				}

				log.Printf("received: \n\t%s\n", string(msg))

				var cmd tictacstatemachine.CommandOneOf
				err = json.Unmarshal(msg, &cmd)
				if err != nil {
					log.Printf("unmarshal error: %v", err)
					continue
				}

				log.Printf("cmd: \n\t%#v\n", cmd)

				game.Handle(cmd.Unwrap())

				if err := game.LastErr(); err != nil {
					log.Printf("game handle command error: %v", err)
				}

				state := game.State()
				stateOneOf := tictacstatemachine.MapStateToOneOf(state)
				result, err := json.Marshal(stateOneOf)
				if err != nil {
					log.Printf("marshal error: %v", err)
					continue
				}

				connections.Range(func(key, value any) bool {
					conn := key.(net.Conn)
					err := wsutil.WriteServerMessage(conn, op, result)
					if err != nil {
						connections.Delete(conn)
						conn.Close()
						log.Printf("write error: %v", err)
					}
					return true

				})
				//err = wsutil.WriteServerMessage(conn, op, result)
				//if err != nil {
				//	handle error
				//}
			}
		}()
	})

	handler := cors.AllowAll().Handler(mux)
	http.ListenAndServe(":8080", handler)
}
