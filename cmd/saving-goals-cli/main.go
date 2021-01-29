package main

import (
	"os"
	"time"

	"github.com/eventually-rs/saving-goals-go/pkg/must"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "create-account",
				Usage:  "sends a AccountCreated message on the Kafka client specified in KAFKA_HOST",
				Action: createAccount,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "account-id",
						Required: true,
						Usage:    "account identifier",
					},
				},
			},
			{
				Name:   "record-account-transaction",
				Usage:  "sends a AccountTransactionRecorded message on the Kafka client specified in KAFKA_HOST",
				Action: recordAccountTransaction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "account-id",
						Required: true,
						Usage:    "account identifier",
					},
					&cli.Float64Flag{
						Name:     "amount",
						Required: true,
						Usage:    "transaction amount",
					},
					&cli.TimestampFlag{
						Name:   "recorded-at",
						Usage:  "timestamp of when the event occurred",
						Layout: time.RFC3339,
					},
				},
			},
		},
	}

	must.NotFail(app.Run(os.Args))
}
