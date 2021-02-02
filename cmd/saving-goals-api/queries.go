package main

import (
	"context"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"

	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/extension/correlation"
	"github.com/eventually-rs/eventually-go/projection"
	"github.com/eventually-rs/eventually-go/subscription"
	"github.com/eventually-rs/eventually-go/subscription/checkpoint"
	"go.uber.org/zap"
)

func buildAccountsWithSavingGoalsReadModel(
	ctx context.Context,
	accountEventStore eventstore.Typed,
	logger *zap.Logger,
) (*account.WithSavingGoalsProjection, error) {
	accountsWithSavingGoals := account.NewWithSavingGoalsProjection()

	accountsWithSavingGoalsSubscription := subscription.CatchUp{
		SubscriptionName: "accounts-with-saving-goals",
		EventStore:       accountEventStore,
		Checkpointer:     checkpoint.NopCheckpointer,
	}

	go func() {
		logger.Info("account.WithSavingGoals projector started")

		accountsWithSavingGoals := correlation.WrapProjection(accountsWithSavingGoals)
		projector := projection.NewProjector(accountsWithSavingGoals, accountsWithSavingGoalsSubscription)

		if err := projector.Start(ctx); err != nil {
			logger.Error("account.WithSavingGoals projector exited with error", zap.Error(err))
		}
	}()

	return accountsWithSavingGoals, nil
}
