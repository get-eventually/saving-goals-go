package consumer

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/resources/messages"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type AccountCreatedMessage struct {
	AccountID string `json:"accountId"`
}

type AccountCreated struct {
	kafkaReader *kafka.Reader
	commandBus  command.Dispatcher
	logger      *zap.Logger
}

func NewAccountCreated(kafkaURL string, commandBus command.Dispatcher, logger *zap.Logger) AccountCreated {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		GroupID: "account-creation-consumer",
		Topic:   "account-creation",
	})

	return AccountCreated{
		kafkaReader: kafkaReader,
		commandBus:  commandBus,
		logger:      logger,
	}
}

func (c AccountCreated) Close() error { return c.kafkaReader.Close() }

func (c AccountCreated) Start(ctx context.Context) error {
	for {
		msg, err := c.kafkaReader.FetchMessage(ctx)

		if errors.Is(err, io.EOF) {
			c.logger.Info("EOF received, closing consumer")
			return nil
		}

		if err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to read message from kafka: %w", err)
		}

		c.logger.Debug("Message received",
			zap.Binary("key", msg.Key),
			zap.Binary("value", msg.Value))

		if err := c.handle(ctx, msg); err != nil {
			c.logger.Error("Failed to handle message", zap.Error(err))
			continue
		}

		if err := c.kafkaReader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("Failed to commit message", zap.Error(err))
		}
	}
}

func (c AccountCreated) handle(ctx context.Context, msg kafka.Message) error {
	var message messages.AccountCreated

	if err := proto.Unmarshal(msg.Value, &message); err != nil {
		if err != nil {
			return fmt.Errorf("consumer.AccountCreated: failed to unmarshal message: %w", err)
		}
	}

	err := c.commandBus.Dispatch(ctx, eventually.Command{
		Payload: account.CreateCommand{
			AccountID: message.AccountId,
		},
		Metadata: eventually.Metadata{
			"Recorded-At": message.RecordedAt,
		},
	})

	if err != nil {
		return fmt.Errorf("consumer.AccountCreated: failed to dispatch command: %w", err)
	}

	return nil
}
