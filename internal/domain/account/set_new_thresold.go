package account

import (
	"context"
	"fmt"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

// SetNewThreshold is the Domain Command used to set a new Threshold value
// in an Account's Saving Goal.
type SetNewThreshold struct {
	AccountID aggregate.StringID
	Value     float64
}

// SetNewThresholdCommandHandler is the Command Handler for SetNewThreshold commands.
type SetNewThresholdCommandHandler struct {
	Repository *aggregate.Repository
}

// CommandType returns a SetNewThreshold instance to bind to this Handler.
func (SetNewThresholdCommandHandler) CommandType() command.Command { return SetNewThreshold{} }

// Handle sets the new Threshold in the Account's Saving Goal with the value
// specified in the Command dispatched.
func (h SetNewThresholdCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(SetNewThreshold)

	account, err := h.Repository.Get(ctx, command.AccountID)
	if err != nil {
		return fmt.Errorf("account.SetNewThresholdCommandHandler: failed to get account: %w", err)
	}

	if err := account.(*Account).SetNewThreshold(command.Value); err != nil {
		return fmt.Errorf("account.SetNewThresholdCommandHandler: failed to set new threshold: %w", err)
	}

	if err := h.Repository.Add(ctx, account); err != nil {
		return fmt.Errorf("account.SetNewThresholdCommandHandler: failed to save new account state: %w", err)
	}

	return nil
}
