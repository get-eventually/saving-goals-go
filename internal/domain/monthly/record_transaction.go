package monthly

import (
	"context"
	"fmt"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

type RecordTransaction struct {
	ID
	Amount float64
}

type RecordTransactionCommandHandler struct {
	Repository *aggregate.Repository
}

func (RecordTransactionCommandHandler) CommandType() command.Command { return RecordTransaction{} }

func (h RecordTransactionCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(RecordTransaction)

	monthlySpending, err := h.Repository.Get(ctx, command.ID)
	if err != nil {
		return fmt.Errorf("monthly.RecordTransaction: failed to get spending aggregate from repository: %w", err)
	}

	if err := monthlySpending.(*Spending).RecordTransaction(command.Amount); err != nil {
		return fmt.Errorf("monthly.RecordTransaction: failed to record transaction in spending: %w", err)
	}

	if err := h.Repository.Add(ctx, monthlySpending); err != nil {
		return fmt.Errorf("monthly.RecordTransaction: failed to save spending status to repository: %w", err)
	}

	return nil
}
