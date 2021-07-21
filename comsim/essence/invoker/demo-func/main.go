package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("Hello from main!")

	go func() {
		<-time.After(time.Second * 2)
		fmt.Println("Exiting container after 5s")
		os.Exit(0)
	}()

	http.HandleFunc("/invoke", func(writer http.ResponseWriter, request *http.Request) {
		in, _ := ioutil.ReadAll(request.Body)
		fmt.Fprintf(writer, "Hello %s, from Docker!", in)
	})
	http.ListenAndServe(":9666", http.DefaultServeMux)
}
