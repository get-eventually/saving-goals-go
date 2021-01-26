package main

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/monthly"

	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/extension/correlation"
	"github.com/eventually-rs/eventually-go/projection"
	"github.com/eventually-rs/eventually-go/subscription"
	"go.uber.org/zap"
)

func startCreateSpendingStartOfTheMonthPolicy(
	ctx context.Context,
	commandBus command.Dispatcher,
	queryBus monthly.QueryDispatcher,
	eventStore eventstore.Store,
	checkpointer subscription.Checkpointer,
	logger *zap.Logger,
) error {
	createSpendingStartOfTheMonthPolicy := monthly.CreateSpendingStartOfTheMonthPolicy{
		CommandDispatcher: commandBus,
		QueryDispatcher:   queryBus,
	}

	createSpendingStartOfTheMonthSubscription, err := subscription.NewCatchUp(
		"create-spending-start-of-the-month",
		eventStore,
		eventStore,
		checkpointer,
	)

	if err != nil {
		return fmt.Errorf("startCreateSpendingStartOfTheMonthPolicy: failed to start subscription: %w", err)
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
