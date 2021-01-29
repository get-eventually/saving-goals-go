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

func recordAccountTransaction(ctx *cli.Context) error {
	config, err := app.ParseConfig()
	if err != nil {
		return fmt.Errorf("recordAccountTransaction: %w", err)
	}

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{config.Kafka.Addr()},
		Topic:   "account-transactions",
	})

	accountID := ctx.String("account-id")
	amount := ctx.Float64("amount")
	recordedAt := ctx.Timestamp("recorded-at")

	recordedTimestamp := timestamppb.Now()
	if recordedAt != nil {
		recordedTimestamp = timestamppb.New(*recordedAt)
	}

	msg, err := proto.Marshal(&messages.AccountTransactionRecorded{
		AccountId:  accountID,
		Amount:     float32(amount),
		RecordedAt: recordedTimestamp,
	})

	if err != nil {
		return fmt.Errorf("recordAccountTransaction: failed to marshal message to protobuf: %w", err)
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(accountID),
		Value: msg,
	})

	if err != nil {
		err = fmt.Errorf("recordAccountTransaction: failed to write message to kafka: %w", err)
	}

	return err
}
