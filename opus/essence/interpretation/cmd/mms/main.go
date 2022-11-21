package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	var app *cli.App
	app = &cli.App{
		Name:                   "mms",
		Description:            "Workflow for mms",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Flags:                  []cli.Flag{},
		Action: func(c *cli.Context) error {
			cwd, _ := syscall.Getwd()
			sourceName := path.Base(os.Getenv("GOFILE"))
			sourcePath := path.Join(cwd, sourceName)

			baseName := strings.TrimSuffix(sourceName, path.Ext(sourceName))

			fmt.Println("sourceName", sourceName)
			fmt.Println("sourcePath", sourcePath)
			fmt.Println("baseName", baseName)
			return nil
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
