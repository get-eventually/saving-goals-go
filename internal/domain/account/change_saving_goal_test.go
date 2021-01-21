package account_test

import (
	"testing"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/scenario"
)

func TestChangeSavingGoal(t *testing.T) {
	t.Run("command fails when the account specified in the command does not exist", func(t *testing.T) {
		scenario.
			CommandHandler().
			When(eventually.Command{
				Payload: account.ChangeSavingGoal{
					AccountID: "test-account",
					SavingGoal: saving.Goal{
						Amount:     500,
						Thresholds: []float64{0.25, 0.5, 0.75, 1},
					},
				},
			}).
			ThenFails().
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.ChangeSavingGoalCommandHandler{Repository: r}
			})
	})

	t.Run("command fails when the new saving goal has no thresholds", func(t *testing.T) {
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
				Payload: account.ChangeSavingGoal{
					AccountID:  "test-account",
					SavingGoal: saving.Goal{Amount: 500},
				},
			}).
			ThenError(account.ErrAtLeastOneThreshold).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.ChangeSavingGoalCommandHandler{Repository: r}
			})
	})

	t.Run("command fails when the saving goal amount is zero", func(t *testing.T) {
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
				Payload: account.ChangeSavingGoal{
					AccountID: "test-account",
					SavingGoal: saving.Goal{
						Amount:     0,
						Thresholds: []float64{0.25, 0.5, 0.75, 1},
					},
				},
			}).
			ThenError(account.ErrGoalIsZero).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.ChangeSavingGoalCommandHandler{Repository: r}
			})
	})

	t.Run("new saving goal with at least one threshold is saved for an existing account", func(t *testing.T) {
		accountID := "test-account"
		newSavingGoal := saving.Goal{
			Amount:     500,
			Thresholds: []float64{0.25, 0.5, 0.75, 1},
		}

		scenario.
			CommandHandler().
			Given(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: accountID,
				Version:    1,
				Event: eventually.Event{
					Payload: account.WasCreated{
						AccountID: accountID,
					},
				},
			}).
			When(eventually.Command{
				Payload: account.ChangeSavingGoal{
					AccountID:  aggregate.StringID(accountID),
					SavingGoal: newSavingGoal,
				},
			}).
			Then(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: accountID,
				Version:    2,
				Event: eventually.Event{
					Payload: account.SavingGoalWasChanged{
						SavingGoal: newSavingGoal,
					},
				},
			}).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.ChangeSavingGoalCommandHandler{Repository: r}
			})
	})
}
