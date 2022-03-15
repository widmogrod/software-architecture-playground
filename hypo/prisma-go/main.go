package main

import (
	"context"
	"encoding/json"
	"fmt"
	db "github.com/widmogrod/software-architecture-playground/hypo/prisma-go/sdk"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		return err
	}

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()

	http.HandleFunc("/create", func(r http.ResponseWriter, rq *http.Request) {
		// create a post
		createdPost, err := client.Post.CreateOne(
			db.Post.Title.Set("Hi from Prisma!"),
			db.Post.Published.Set(true),
			db.Post.Desc.Set("Prisma is a database toolkit and makes databases easy."),
		).Exec(ctx)
		if err != nil {
			reportErr(r, err)
			return
		}

		result, err := json.MarshalIndent(createdPost, "", "  ")
		if err != nil {
			reportErr(r, err)
			return
		}

		r.Header().Set("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)
		r.Write(result)
	})

	http.HandleFunc("/", func(r http.ResponseWriter, rq *http.Request) {
		// create a post
		createdPost, err := client.Post.FindMany().Exec(ctx)
		if err != nil {
			reportErr(r, err)
			return
		}

		result, err := json.MarshalIndent(createdPost, "", "  ")
		if err != nil {
			reportErr(r, err)
			return
		}

		r.Header().Set("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)
		r.Write(result)
	})

	// find a single post
	http.HandleFunc("/get", func(r http.ResponseWriter, rq *http.Request) {
		postID := rq.URL.Query().Get("id")
		post, err := client.Post.FindUnique(
			db.Post.ID.Equals(postID),
		).Exec(ctx)
		if err != nil {
			reportErr(r, err)
			return
		}

		result, err := json.MarshalIndent(post, "", "  ")
		if err != nil {
			reportErr(r, err)
			return
		}

		r.Header().Set("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)
		r.Write(result)
	})

	log.Println("port: " + port())
	http.ListenAndServe(port(), http.DefaultServeMux)

	return nil
}

func port() string {
	r, found := os.LookupEnv("APP_PORT")
	if found {
		return ":" + r
	}

	return ":8080"
}

func reportErr(r http.ResponseWriter, err error) {
	r.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(r, "err=%v", err)
}
