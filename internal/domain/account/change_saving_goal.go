package account

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

// ChangeSavingGoal is the Domain Command used to change the Saving Goal
// settings of an Account.
type ChangeSavingGoal struct {
	AccountID  aggregate.StringID
	SavingGoal saving.Goal
}

// ChangeSavingGoalCommandHandler is the Command Handler for ChangeSavingGoal commands.
type ChangeSavingGoalCommandHandler struct {
	Repository *aggregate.Repository
}

// CommandType returns a ChangeSavingGoal instance to bind to this Handler.
func (ChangeSavingGoalCommandHandler) CommandType() command.Command { return ChangeSavingGoal{} }

// Handle changes the Saving Goal of an Account with the one specified in the
// Command dispatched.
func (h ChangeSavingGoalCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(ChangeSavingGoal)

	account, err := h.Repository.Get(ctx, command.AccountID)
	if err != nil {
		return fmt.Errorf("account.ChangeSavingGoalCommandHandler: failed to get account: %w", err)
	}

	if err := account.(*Account).ChangeSavingGoal(command.SavingGoal); err != nil {
		return fmt.Errorf("account.ChangeSavingGoalCommandHandler: failed to change saving goal: %w", err)
	}

	if err := h.Repository.Add(ctx, account); err != nil {
		return fmt.Errorf("account.ChangeSavingGoalCommandHandler: failed to save new account state: %w", err)
	}

	return nil
}
