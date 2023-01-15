package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/rs/cors"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"io"
	"log"
	"net/http"
)

func main() {
	reg := storage.NewRepositoryInMemory(tictactoemanage.NewMachine)
	proto := websockproto.NewProtocol[tictactoemanage.Command, tictactoemanage.State](reg)

	proto.UnmarshalCommand = func(msg []byte) (tictactoemanage.Command, error) {
		sch, err := schema.FromJSON(msg)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalCommand: %s", err)
		}

		goo := schema.ToGo(sch)

		cmd, ok := goo.(tictactoemanage.Command)
		if !ok {
			return nil, fmt.Errorf("UnmarshalCommand: %T not a command", goo)
		}

		return cmd, nil
	}
	proto.MarshalState = func(state tictactoemanage.State) ([]byte, error) {
		result := schema.FromGo(state)
		return schema.ToJSON(result)
	}
	proto.ExtractSessionID = func(cmd tictactoemanage.Command) string {
		return tictactoemanage.MustMatchCommand(
			cmd,
			func(x *tictactoemanage.CreateSessionCMD) string {
				return x.SessionID
			},
			func(x *tictactoemanage.JoinGameSessionCMD) string {
				return x.SessionID
			},
			func(x *tictactoemanage.GameSessionWithBotCMD) string {
				return x.SessionID
			},
			func(x *tictactoemanage.LeaveGameSessionCMD) string {
				return x.SessionID
			},
			func(x *tictactoemanage.NewGameCMD) string {
				return x.SessionID
			},
			func(x *tictactoemanage.GameActionCMD) string {
				return x.SessionID
			},
		)
	}

	go proto.PublishLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	mux.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		proto.ConnectionOpen(conn)

		go func() {
			defer proto.ConnectionClose(conn)
			defer conn.Close()

			for {
				err = proto.ConnectionReceiveData(conn)
				if err != nil {
					if err == io.EOF {
						log.Println("ConnectionReceiveData: CLOSED")
						break
					}
					log.Println("ConnectionReceiveData:", err)
					continue
				}
			}
		}()
	})

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
