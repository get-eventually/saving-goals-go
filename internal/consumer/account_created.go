package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/segmentio/kafka-go"
)

type AccountCreatedMessage struct {
	AccountID string `json:"accountId"`
}

type AccountCreated struct {
	kafkaReader *kafka.Reader
	commandBus  command.Dispatcher
}

func NewAccountCreated(kafkaURL string, commandBus command.Dispatcher) AccountCreated {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		GroupID: "account-creation-consumer",
		Topic:   "account-creation",
	})

	return AccountCreated{
		kafkaReader: kafkaReader,
		commandBus:  commandBus,
	}
}

func (c AccountCreated) Close() error { return c.kafkaReader.Close() }

func (c AccountCreated) Start(ctx context.Context) error {
	for {
		msg, err := c.kafkaReader.FetchMessage(ctx)

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to read message from kafka: %w", err)
		}

		var message AccountCreatedMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to unmarshal message: %w", err)
		}

		err = c.commandBus.Dispatch(ctx, eventually.Command{
			Payload: account.CreateCommand{
				AccountID: message.AccountID,
			},
		})

		if err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to dispatch command: %w", err)
		}

		if err := c.kafkaReader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to commit message: %w", err)
		}
	}
}
