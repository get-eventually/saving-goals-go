package account_test

import (
	"testing"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/scenario"
)

func TestCreateAccount(t *testing.T) {
	t.Run("new account is created if it was not created before", func(t *testing.T) {
		scenario.
			CommandHandler().
			When(eventually.Command{
				Payload: account.CreateCommand{
					AccountID: "test-account",
				},
			}).
			Then(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: "test-account",
				Version:    1,
				Event: eventually.Event{
					Payload: account.WasCreated{
						AccountID: "test-account",
					},
				},
			}).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.CreateCommandHandler{Repository: r}
			})
	})

	t.Run("create an account with already existing id fails", func(t *testing.T) {
		scenario.
			CommandHandler().
			Given(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: "test-account",
				Version:    1,
				Event: eventually.Event{
					Payload: account.WasCreated{
						AccountID: "test-account",
					},
				},
			}).
			When(eventually.Command{
				Payload: account.CreateCommand{
					AccountID: "test-account",
				},
			}).
			ThenFails().
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.CreateCommandHandler{Repository: r}
			})
	})
}
