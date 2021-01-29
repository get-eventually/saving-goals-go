package consumer

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/resources/messages"
	"google.golang.org/protobuf/proto"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type AccountTransactionRecorded struct {
	kafkaReader *kafka.Reader
	deadLetter  *kafka.Writer
	commandBus  command.Dispatcher
	logger      *zap.Logger
}

func NewAccountTransactionRecorded(
	kafkaURL string,
	commandBus command.Dispatcher,
	logger *zap.Logger,
) AccountTransactionRecorded {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		GroupID: "account-transactions-consumer",
		Topic:   "account-transactions",
	})

	deadLetterWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaURL},
		Topic:   "saving-goals.account-transactions.dead",
	})

	return AccountTransactionRecorded{
		kafkaReader: kafkaReader,
		deadLetter:  deadLetterWriter,
		commandBus:  commandBus,
		logger:      logger,
	}
}

func (c AccountTransactionRecorded) Close() error { return c.kafkaReader.Close() }

func (c AccountTransactionRecorded) Start(ctx context.Context) error {
	for {
		msg, err := c.kafkaReader.FetchMessage(ctx)

		if errors.Is(err, io.EOF) {
			c.logger.Info("EOF received, closing consumer")
			return nil
		}

		if err != nil {
			return fmt.Errorf("consumer.AccountTransactionRecorded: failed to read message from kafka: %w", err)
		}

		c.logger.Debug("Message received",
			zap.Binary("key", msg.Key),
			zap.Binary("value", msg.Value))

		if err := c.handle(ctx, msg); err != nil {
			c.logger.Warn("Failed to handle message", zap.Error(err))

			err = c.deadLetter.WriteMessages(ctx, kafka.Message{
				Key:   msg.Key,
				Value: msg.Value,
			})

			if err != nil {
				c.logger.Error("Failed to deadletter message", zap.Error(err))
				continue
			}
		}

		if err := c.kafkaReader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("Failed to commit message", zap.Error(err))
		}
	}
}

func (c AccountTransactionRecorded) handle(ctx context.Context, msg kafka.Message) error {
	var message messages.AccountTransactionRecorded

	if err := proto.Unmarshal(msg.Value, &message); err != nil {
		if err != nil {
			return fmt.Errorf("consumer.AccountTransactionRecorded: failed to unmarshal message: %w", err)
		}
	}

	err := c.commandBus.Dispatch(ctx, eventually.Command{
		Payload: account.RecordTransaction{
			AccountID:  aggregate.StringID(message.AccountId),
			Amount:     float64(message.Amount),
			RecordedAt: message.RecordedAt.AsTime(),
		},
	})

	if err != nil {
		return fmt.Errorf("consumer.AccountTransactionRecorded: failed to dispatch command: %w", err)
	}

	return nil
}
