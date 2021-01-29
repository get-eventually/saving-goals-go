package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eventually-rs/saving-goals-go/internal/app"
	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"github.com/eventually-rs/saving-goals-go/internal/domain/monthly"
	"github.com/eventually-rs/saving-goals-go/internal/httpapi"
	"github.com/eventually-rs/saving-goals-go/pkg/must"
	"github.com/eventually-rs/saving-goals-go/pkg/shutdown"

	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/eventstore/postgres"
	"github.com/eventually-rs/eventually-go/extension/correlation"
	"github.com/eventually-rs/eventually-go/query"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// <Config> --------------------------------------------------------------------------------------------------------
	config, err := app.ParseConfig()
	must.NotFail(err)
	// </Config> -------------------------------------------------------------------------------------------------------

	// <Logger> --------------------------------------------------------------------------------------------------------
	logger, err := app.NewLogger()
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

	must.NotFail(eventStore.Register(ctx, "month", map[string]interface{}{
		"month_started": interval.MonthStarted{},
	}))

	must.NotFail(eventStore.Register(ctx, account.Type.Name(), map[string]interface{}{
		"account_was_created":              account.WasCreated{},
		"saving_goal_was_changed":          account.SavingGoalWasChanged{},
		"saving_goal_was_disabled":         account.SavingGoalWasDisabled{},
		"threshold_was_set":                account.ThresholdWasSet{},
		"account_transaction_was_recorded": account.TransactionWasRecorded{},
	}))

	must.NotFail(eventStore.Register(ctx, monthly.Type.Name(), map[string]interface{}{
		"monthly_spending_tracking_started":         monthly.SpendingTrackingStarted{},
		"monthly_spending_transaction_was_recorded": monthly.TransactionWasRecorded{},
		"monthly_spending_limit_was_updated":        monthly.SpendingLimitWasUpdated{},
		"monthly_spending_threshold_was_reached":    monthly.ThresholdWasReached{},
	}))

	monthEventStore, err := eventStore.Type(ctx, "month")
	must.NotFail(err)

	accountEventStore, err := eventStore.Type(ctx, account.Type.Name())
	must.NotFail(err)

	monthlySpendingEventStore, err := eventStore.Type(ctx, monthly.Type.Name())
	must.NotFail(err)

	checkpointer := postgresEventStore
	// </EventStore> ---------------------------------------------------------------------------------------------------

	// <Repositories> --------------------------------------------------------------------------------------------------
	accountRepository := aggregate.NewRepository(account.Type, accountEventStore)
	monthlySpendingRepository := aggregate.NewRepository(monthly.Type, monthlySpendingEventStore)
	// </Repositories> -------------------------------------------------------------------------------------------------

	// <Queries> -------------------------------------------------------------------------------------------------------
	queryBus := query.NewSimpleBus()

	accountsWithSavingGoals, err := buildAccountsWithSavingGoalsReadModel(ctx, accountEventStore, logger)
	must.NotFail(err)

	queryBus.Register(accountsWithSavingGoals)
	// </Queries> ------------------------------------------------------------------------------------------------------

	// <Commands> ------------------------------------------------------------------------------------------------------
	commandBus := command.NewSimpleBus()

	commandBus.Register(account.CreateCommandHandler{Repository: accountRepository})
	commandBus.Register(account.ChangeSavingGoalCommandHandler{Repository: accountRepository})
	commandBus.Register(account.SetNewThresholdCommandHandler{Repository: accountRepository})
	commandBus.Register(account.RecordTransactionCommandHandler{Repository: accountRepository})

	commandBus.Register(monthly.StartSpendingTrackingCommandHandler{Repository: monthlySpendingRepository})
	commandBus.Register(monthly.RecordTransactionCommandHandler{Repository: monthlySpendingRepository})
	// </Commands> -----------------------------------------------------------------------------------------------------

	// <ProcessManagers> -----------------------------------------------------------------------------------------------
	must.NotFail(startCreateSpendingStartOfTheMonthPolicy(ctx, commandBus, queryBus, eventStore, checkpointer, logger))
	must.NotFail(startRecordTransactionPolicy(ctx, commandBus, queryBus, accountEventStore, checkpointer, logger))
	// </ProcessManagers> ----------------------------------------------------------------------------------------------

	// <HttpServer> ----------------------------------------------------------------------------------------------------
	router := httpapi.NewRouter(commandBus, monthEventStore, logger)

	httpServer := &http.Server{
		Addr:    config.Server.Addr(),
		Handler: router,
	}

	logger.Info("Server started",
		zap.Uint16("port", config.Server.Port),
		zap.String("addr", fmt.Sprintf("http://%s", config.Server.Addr())))

	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server closed with an error", zap.Error(err))
		} else {
			logger.Info("Server shutdown successfully")
		}
	}()

	<-shutdown.Gracefully()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	logger.Info("Shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown returned an error", zap.Error(err))
	}
	// </HttpServer> ---------------------------------------------------------------------------------------------------
}
