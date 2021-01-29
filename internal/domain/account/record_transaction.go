package account

import (
	"context"
	"fmt"
	"time"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

// RecordTransaction is the Domain Command used to record a new Transaction
// involving the specified Account.
type RecordTransaction struct {
	AccountID  aggregate.StringID
	Amount     float64
	RecordedAt time.Time
}

// RecordTransactionCommandHandler is the Command Handler for RecordTransaction commands.
type RecordTransactionCommandHandler struct {
	Repository *aggregate.Repository
}

// CommandType returns a new RecordTransaction instance to bind to this Handler.
func (RecordTransactionCommandHandler) CommandType() command.Command { return RecordTransaction{} }

// Handle records the new transaction amount by updating the Account's balance.
func (h RecordTransactionCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(RecordTransaction)

	account, err := h.Repository.Get(ctx, command.AccountID)
	if err != nil {
		return fmt.Errorf("account.RecordTransaction: failed to get account: %w", err)
	}

	if err := account.(*Account).RecordTransaction(command.Amount, command.RecordedAt); err != nil {
		return fmt.Errorf("account.RecordTransaction: failed to record transaction: %w", err)
	}

	if err := h.Repository.Add(ctx, account); err != nil {
		return fmt.Errorf("account.RecordTransaction: failed to save new account state: %w", err)
	}

	return nil
}
