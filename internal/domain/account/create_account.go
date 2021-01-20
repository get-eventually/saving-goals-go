package account

import (
	"context"
	"fmt"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
)

var _ command.Handler = CreateCommandHandler{}

// CreateCommand is the Domain Command to create new Accounts.
type CreateCommand struct {
	AccountID string
}

// CreateCommandHandler is the Command Handler for CreateCommand messages.
type CreateCommandHandler struct {
	Repository *aggregate.Repository
}

// CommandType returns a CreateCommand instance to bind to this command type.
func (CreateCommandHandler) CommandType() command.Command { return CreateCommand{} }

// Handle handles a CreateCommand message, by trying to create a new Account
// and saving it into the data store.
func (h CreateCommandHandler) Handle(ctx context.Context, cmd eventually.Command) error {
	command := cmd.Payload.(CreateCommand)

	account, err := Create(command.AccountID)
	if err != nil {
		return fmt.Errorf("account.CreateCommandHandler: failed to create new account: %w", err)
	}

	if err := h.Repository.Add(ctx, account); err != nil {
		return fmt.Errorf("account.CreateCommandHandler: failed to save account into repository: %w", err)
	}

	return nil
}
