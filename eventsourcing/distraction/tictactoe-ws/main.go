package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/cors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type ma = map[string]interface{}

func main() {
	registry := sync.Map{}
	peers := sync.Map{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	mux.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/play/")
		fmt.Println("session uuid", uuid)
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		go func() {
			defer conn.Close()

			connectionsAny, _ := peers.LoadOrStore(uuid, &sync.Map{})
			connections := connectionsAny.(*sync.Map)
			connections.Store(conn, conn)
			defer connections.Delete(conn)

			manageAny, _ := registry.LoadOrStore(uuid, tictactoemanage.NewMachine())
			manage := manageAny.(*tictactoemanage.Machine)

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

				var o map[string]interface{}
				err = json.Unmarshal(msg, &o)
				if err != nil {
					log.Printf("unmarshal error: %v", err)
					continue
				}

				var action []byte
				if o1, ok := o["GameActionCMD"]; ok {
					if o2, ok := o1.(ma)["Action"]; ok {
						action, _ = json.Marshal(o2.(ma))
						delete(o2.(ma), "Action")
						msg, _ = json.Marshal(o)
					}
				}

				var cmd tictactoemanage.CommandOneOf
				err = json.Unmarshal(msg, &cmd)
				if err != nil {
					log.Printf("unmarshal error: %v", err)
					continue
				}

				if action != nil {
					var a tictacstatemachine.CommandOneOf

					_ = json.Unmarshal(action, &a)
					cmd.GameActionCMD.Action = tictacstatemachine.UnwrapCommandOneOf(&a)
				}

				log.Printf("cmd: \n\t%#v\n", cmd)
				log.Printf("GameActionCMD:: \n\t%#v\n", cmd.GameActionCMD)
				if cmd.GameActionCMD != nil {
					log.Printf("Action:: \n\t%#v\n", cmd.GameActionCMD.Action)
				}

				manage.Handle(cmd.Unwrap())
				if err := manage.LastErr(); err != nil {
					log.Printf("game handle command error: %v", err)
				}

				state := manage.State()
				log.Printf("State:: \n\t%#v\n", state)
				stateOneOf := tictactoemanage.WrapStateOneOf(state)
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
			}
		}()
	})

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
