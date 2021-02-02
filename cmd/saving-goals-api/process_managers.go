package main

import (
	"context"

	"github.com/eventually-rs/saving-goals-go/internal/domain/monthly"

	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/extension/correlation"
	"github.com/eventually-rs/eventually-go/projection"
	"github.com/eventually-rs/eventually-go/subscription"
	"github.com/eventually-rs/eventually-go/subscription/checkpoint"
	"go.uber.org/zap"
)

func startCreateSpendingStartOfTheMonthPolicy(
	ctx context.Context,
	commandBus command.Dispatcher,
	queryBus monthly.QueryDispatcher,
	eventStore eventstore.Store,
	checkpointer checkpoint.Checkpointer,
	logger *zap.Logger,
) error {
	createSpendingStartOfTheMonthPolicy := monthly.CreateSpendingStartOfTheMonthPolicy{
		CommandDispatcher: commandBus,
		QueryDispatcher:   queryBus,
	}

	createSpendingStartOfTheMonthSubscription := subscription.CatchUp{
		SubscriptionName: "create-spending-start-of-the-month",
		Checkpointer:     checkpointer,
		EventStore:       eventStore,
	}

	go func() {
		logger.Info("monthly.CreateSpendingStartOfTheMonthPolicy projector started")

		createSpendingStartOfTheMonthPolicy := correlation.WrapProjection(createSpendingStartOfTheMonthPolicy)
		projector := projection.NewProjector(
			createSpendingStartOfTheMonthPolicy,
			createSpendingStartOfTheMonthSubscription,
		)

		if err := projector.Start(ctx); err != nil {
			logger.Error("monthly.CreateSpendingStartOfTheMonthPolicy projector exited with error", zap.Error(err))
		}
	}()

	return nil
}

func startRecordTransactionPolicy(
	ctx context.Context,
	commandBus command.Dispatcher,
	queryBus monthly.QueryDispatcher,
	accountStore eventstore.Typed,
	checkpointer checkpoint.Checkpointer,
	logger *zap.Logger,
) error {
	recordTransactionPolicy := monthly.RecordTransactionPolicy{
		CommandDispatcher: commandBus,
		Logger:            logger,
	}

	recordTransactionSubscription := subscription.CatchUp{
		SubscriptionName: "record-transaction",
		EventStore:       accountStore,
		Checkpointer:     checkpointer,
	}

	go func() {
		logger.Info("monthly.RecordTransactionPolicy projector started")

		recordTransactionPolicy := correlation.WrapProjection(recordTransactionPolicy)
		projector := projection.NewProjector(
			recordTransactionPolicy,
			recordTransactionSubscription,
		)

		if err := projector.Start(ctx); err != nil {
			logger.Error("monthly.RecordTransactionPolicy projector exited with error", zap.Error(err))
		}
	}()

	return nil
}
