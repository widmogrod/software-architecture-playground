package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/cors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/schema"
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

type Repository[A any] struct {
	store sync.Map
	new   func() A
}

var ErrNotFound = fmt.Errorf("not found")

func (r *Repository[A]) Get(key string) (A, error) {
	v, ok := r.store.Load(key)
	if !ok {
		var a A
		return a, ErrNotFound
	}
	return v.(A), nil
}

func (r *Repository[A]) Set(key string, value A) error {
	r.store.Store(key, value)
	return nil
}

func (r *Repository[A]) GetOrNew(s string) (A, error) {
	v, err := r.Get(s)
	if err == nil {
		return v, nil
	}

	if err != nil && err != ErrNotFound {
		var a A
		return a, err
	}

	v = r.new()

	err = r.Set(s, v)
	if err != nil {
		var a A
		return a, err
	}

	return v, nil
}

func main() {

	reg := Repository[*tictactoemanage.Machine]{
		store: sync.Map{},
		new: func() *tictactoemanage.Machine {
			return tictactoemanage.NewMachine()
		},
		//do: func(m *tictactoemanage.Machine, cmd tictactoemanage.Command) error {
		//	m.Handle(cmd)
		//},
	}
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

			manage, err := reg.GetOrNew(uuid)
			if err != nil {
				panic(err)
			}

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

				sch, err := schema.JsonToSchema(msg)
				if err != nil {
					log.Printf("JsonToSchema: %v", err)
					continue
				}

				goo := schema.SchemaToGo(
					sch,
					schema.WhenPath(
						[]string{},
						schema.UseStruct(&tictactoemanage.CommandOneOf{}),
					),
					schema.WhenPath(
						[]string{"CreateSessionCMD"},
						schema.UseStruct(&tictactoemanage.CreateSessionCMD{}),
					),
					schema.WhenPath(
						[]string{"JoinGameSessionCMD"},
						schema.UseStruct(&tictactoemanage.JoinGameSessionCMD{}),
					),
					schema.WhenPath(
						[]string{"GameSessionWithBotCMD"},
						schema.UseStruct(&tictactoemanage.GameSessionWithBotCMD{}),
					),
					schema.WhenPath(
						[]string{"NewGameCMD"},
						schema.UseStruct(&tictactoemanage.NewGameCMD{}),
					),
					schema.WhenPath(
						[]string{"LeaveGameSessionCMD"},
						schema.UseStruct(&tictactoemanage.LeaveGameSessionCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD"},
						schema.UseStruct(&tictactoemanage.GameActionCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action"},
						schema.UseStruct(&tictacstatemachine.CommandOneOf{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action", "CreateGameCMD"},
						schema.UseStruct(&tictacstatemachine.CreateGameCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action", "JoinGameCMD"},
						schema.UseStruct(&tictacstatemachine.JoinGameCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action", "StartGameCMD"},
						schema.UseStruct(&tictacstatemachine.StartGameCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action", "MoveCMD"},
						schema.UseStruct(&tictacstatemachine.MoveCMD{}),
					),
					schema.WhenPath(
						[]string{"GameActionCMD", "Action", "GiveUpCMD"},
						schema.UseStruct(&tictacstatemachine.GiveUpCMD{}),
					),
				)

				cmd, ok := goo.(*tictactoemanage.CommandOneOf)
				if !ok {
					log.Printf("invalid command: %#v", goo)
					continue
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
