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

func TestRecordTransaction(t *testing.T) {
	t.Run("command fails when the account specified in the command does not exist", func(t *testing.T) {
		scenario.
			CommandHandler().
			When(eventually.Command{
				Payload: account.RecordTransaction{
					AccountID: aggregate.StringID("test-account"),
					Amount:    200,
				},
			}).
			ThenFails().
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.RecordTransactionCommandHandler{Repository: r}
			})
	})

	t.Run("accound balance is updated accordingly", func(t *testing.T) {
		accountID := "test-account"

		scenario.
			CommandHandler().
			Given(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: accountID,
				Version:    1,
				Event: eventually.Event{
					Payload: account.WasCreated{AccountID: accountID},
				},
			}, eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: accountID,
				Version:    2,
				Event: eventually.Event{
					Payload: account.TransactionWasRecorded{Amount: 1000},
				},
			}).
			When(eventually.Command{
				Payload: account.RecordTransaction{
					AccountID: aggregate.StringID(accountID),
					Amount:    -200,
				},
			}).
			Then(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: accountID,
				Version:    3,
				Event: eventually.Event{
					Payload: account.TransactionWasRecorded{Amount: -200},
				},
			}).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.RecordTransactionCommandHandler{Repository: r}
			})
	})
}
