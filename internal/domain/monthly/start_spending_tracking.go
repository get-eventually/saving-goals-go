package monthly

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

type StartSpendingTracking struct {
	AccountID       string
	Month           interval.Month
	StartingBalance float64
	SavingGoal      saving.Goal
}

type StartSpendingTrackingCommandHandler struct {
	Repository *aggregate.Repository
}

func (StartSpendingTrackingCommandHandler) CommandType() command.Command {
	return StartSpendingTracking{}
}

func (h StartSpendingTrackingCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(StartSpendingTracking)

	monthlySpending, err := NewSpending(command.AccountID, command.Month, command.StartingBalance, command.SavingGoal)
	if err != nil {
		return fmt.Errorf("monthly.StartSpendingTracking: failed to start new spending tracking: %w", err)
	}

	if err := h.Repository.Add(ctx, monthlySpending); err != nil {
		return fmt.Errorf("monthly.StartSpendingTracking: failed to add spending to repository: %w", err)
	}

	return nil
}
