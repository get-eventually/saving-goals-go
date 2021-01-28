package main

import (
	"context"

	"github.com/eventually-rs/saving-goals-go/internal/consumer"
	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/pkg/must"

	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/eventstore/postgres"
	"github.com/eventually-rs/eventually-go/extension/correlation"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// <Config> --------------------------------------------------------------------------------------------------------
	config, err := ParseConfig()
	must.NotFail(err)
	// </Config> -------------------------------------------------------------------------------------------------------

	// <Logger> --------------------------------------------------------------------------------------------------------
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)

	logger, err := loggerConfig.Build()
	must.NotFail(err)

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}()
	// </Logger> -------------------------------------------------------------------------------------------------------

	// <EventStore> ----------------------------------------------------------------------------------------------------
	postgresEventStore, err := postgres.OpenEventStore(config.Database.DSN())
	must.NotFail(err)

	defer func() {
		if err := postgresEventStore.Close(); err != nil {
			logger.Error("Closing the event store exited with error", zap.Error(err))
		}
	}()

	// Use correlated Event Store to embed additional metadata into appended events.
	eventStore := eventstore.Store(postgresEventStore)
	eventStore = correlation.WrapEventStore(eventStore, func() string {
		return uuid.New().String()
	})

	must.NotFail(eventStore.Register(ctx, account.Type.Name(), map[string]interface{}{
		"account_was_created":              account.WasCreated{},
		"saving_goal_was_changed":          account.SavingGoalWasChanged{},
		"saving_goal_was_disabled":         account.SavingGoalWasDisabled{},
		"threshold_was_set":                account.ThresholdWasSet{},
		"account_transaction_was_recorded": account.TransactionWasRecorded{},
	}))

	accountEventStore, err := eventStore.Type(ctx, account.Type.Name())
	must.NotFail(err)
	// </EventStore> ---------------------------------------------------------------------------------------------------

	// <Repositories> --------------------------------------------------------------------------------------------------
	accountRepository := aggregate.NewRepository(account.Type, accountEventStore)
	// </Repositories> -------------------------------------------------------------------------------------------------

	// <Commands> ------------------------------------------------------------------------------------------------------
	commandBus := command.NewSimpleBus()

	commandBus.Register(account.CreateCommandHandler{Repository: accountRepository})
	commandBus.Register(account.RecordTransactionCommandHandler{Repository: accountRepository})
	// </Commands> -----------------------------------------------------------------------------------------------------

	// <KafkaConsumers> ------------------------------------------------------------------------------------------------
	accountCreatedConsumer := consumer.NewAccountCreated(
		config.Kafka.Addr(),
		commandBus,
	)

	accountTransactionRecordedConsumer := consumer.NewAccountTransactionRecorded(
		config.Kafka.Addr(),
		commandBus,
		logger,
	)

	defer func() {
		if err := accountCreatedConsumer.Close(); err != nil {
			logger.Error("account-creation-consumer failed to close", zap.Error(err))
		}
	}()

	defer func() {
		if err := accountTransactionRecordedConsumer.Close(); err != nil {
			logger.Error("account-transactions-consumer failed to close", zap.Error(err))
		}
	}()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		logger.Info("account-creation-consumer started")
		return accountCreatedConsumer.Start(ctx)
	})

	group.Go(func() error {
		logger.Info("account-transactions-consumer started")
		return accountTransactionRecordedConsumer.Start(ctx)
	})

	if err := group.Wait(); err != nil {
		logger.Fatal("Consumers exited with error", zap.Error(err))
	}
	// </KafkaConsumers> -----------------------------------------------------------------------------------------------
}
