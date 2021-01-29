package main

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/app"
	"github.com/eventually-rs/saving-goals-go/resources/messages"
	"github.com/golang/protobuf/proto"

	"github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createAccount(ctx *cli.Context) error {
	config, err := app.ParseConfig()
	if err != nil {
		return fmt.Errorf("createAccount: %w", err)
	}

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{config.Kafka.Addr()},
		Topic:   "account-creation",
	})

	accountID := ctx.String("account-id")
	if accountID == "" {
		return fmt.Errorf("createAccount: 'account-id' should be specified")
	}

	msg, err := proto.Marshal(&messages.AccountCreated{
		AccountId:  accountID,
		RecordedAt: timestamppb.Now(),
	})

	if err != nil {
		return fmt.Errorf("createAccount: failed to marshal message to protobuf: %w", err)
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(accountID),
		Value: msg,
	})

	if err != nil {
		err = fmt.Errorf("createAccount: failed to write message to kafka: %w", err)
	}

	return err
}
