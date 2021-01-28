package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type AccountTransctionRecordedMessage struct {
	AccountID string  `json:"accountId"`
	Amount    float64 `json:"amount"`
}

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
			c.logger.Error("Failed to read message", zap.Error(err))
			return fmt.Errorf("consumer.AccountTransactionRecorded: failed to read message from kafka: %w", err)
		}

		c.logger.Debug("Message received",
			zap.Binary("key", msg.Key),
			zap.Binary("value", msg.Value))

		var message AccountTransctionRecordedMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			err = c.deadLetter.WriteMessages(ctx, kafka.Message{
				Key:   msg.Key,
				Value: msg.Value,
			})

			if err != nil {
				return fmt.Errorf("consumer.AccountTransactionRecorded: failed to deadletter poisoned message: %w", err)
			}
		}

		err = c.commandBus.Dispatch(ctx, eventually.Command{
			Payload: account.RecordTransaction{
				AccountID: aggregate.StringID(message.AccountID),
				Amount:    message.Amount,
			},
		})

		if err != nil {
			return fmt.Errorf("consumer.AccountTransactionRecorded: failed to dispatch command: %w", err)
		}

		if err := c.kafkaReader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("consumer.AccountTransactionRecorded: failed to commit message: %w", err)
		}
	}
}
