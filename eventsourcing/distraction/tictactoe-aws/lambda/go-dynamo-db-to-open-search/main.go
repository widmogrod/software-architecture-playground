package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      false,
		DisableQuote:     true,
		DisableTimestamp: true,
	})

	di := tictactoe_game_server.DefaultDI(
		tictactoe_game_server.RunAWS,
	)

	repo := di.GetOpenQueryStorage()

	lambda.Start(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		update := schemaless.UpdateRecords[schemaless.Record[schema.Schema]]{
			UpdatingPolicy: schemaless.PolicyOverwriteServerChanges,
			Saving:         map[string]schemaless.Record[schema.Schema]{},
			Deleting:       map[string]schemaless.Record[schema.Schema]{},
		}

		for _, record := range event.Records {
			if record.EventName == "REMOVE" {
				data, err := json.Marshal(record.Change.OldImage)
				if err != nil {
					log.Errorf("Delete(1): %s \n", err)
					continue
				}

				schemed, err := schema.FromJSON(data)
				if err != nil {
					log.Errorf("Delete(2): %s \n", err)
					continue
				}

				typed, err := toTyped(schemed)
				if err != nil {
					log.Errorf("Delete(3): %s \n", err)
					continue
				}

				update.Deleting[typed.ID+typed.Type] = typed
			} else if record.Change.NewImage != nil {
				data, err := json.Marshal(record.Change.NewImage)
				if err != nil {
					log.Errorf("Put(1): %s \n", err)
					continue
				}

				schemed, err := schema.FromJSON(data)
				if err != nil {
					log.Errorf("Put(2): %s \n", err)
					continue
				}

				typed, err := toTyped(schemed)
				if err != nil {
					log.Errorf("Put(3): %s \n", err)
					continue
				}

				update.Saving[typed.ID+typed.Type] = typed
			}
		}

		err := repo.UpdateRecords(update)
		if err != nil {
			log.Errorf("UpdateRecords(1): %s \n", err)
		} else {
			log.Infof("UpdateRecords(2): OK \n")
		}

		return events.DynamoDBEventResponse{}, nil
	})
}

func toTyped(record schema.Schema) (schemaless.Record[schema.Schema], error) {
	normalised, err := schema.UnwrapDynamoDB(record)
	if err != nil {
		data, err := schema.ToJSON(record)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return schemaless.Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream unwrap DynamoDB record: %v", record)
	}

	typed := schemaless.Record[schema.Schema]{
		ID:      schema.AsDefault[string](schema.Get(normalised, "ID"), "record-id-corrupted"),
		Type:    schema.AsDefault[string](schema.Get(normalised, "Type"), "record-id-corrupted"),
		Data:    schema.Get(normalised, "Data"),
		Version: schema.AsDefault[uint16](schema.Get(normalised, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		data, err := schema.ToJSON(normalised)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return schemaless.Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream corrupted record: %v", normalised)
	}
	return typed, nil
}
