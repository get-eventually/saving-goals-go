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

func TestSetNewThreshold(t *testing.T) {
	t.Run("command fails when the account specified in the command does not exist", func(t *testing.T) {
		scenario.
			CommandHandler().
			When(eventually.Command{
				Payload: account.SetNewThreshold{
					AccountID: "test-account",
					Value:     0.35,
				},
			}).
			ThenFails().
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.SetNewThresholdCommandHandler{Repository: r}
			})
	})

	t.Run("command fails when the account has no saving goal set", func(t *testing.T) {
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
				Payload: account.SetNewThreshold{
					AccountID: "test-account",
					Value:     0.35,
				},
			}).
			ThenError(account.ErrNoSavingGoal).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.SetNewThresholdCommandHandler{Repository: r}
			})
	})

	t.Run("command fails when the account saving goal has same threshold already set", func(t *testing.T) {
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
			}, eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: "test-account",
				Version:    2,
				Event: eventually.Event{
					Payload: account.SavingGoalWasChanged{
						SavingGoal: saving.Goal{
							Amount:     500,
							Thresholds: []float64{0.25, 0.5},
						},
					},
				},
			}).
			When(eventually.Command{
				Payload: account.SetNewThreshold{
					AccountID: "test-account",
					Value:     0.5,
				},
			}).
			ThenError(account.ErrThresholdAlreadyExists).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.SetNewThresholdCommandHandler{Repository: r}
			})
	})

	t.Run("new threshold is set if was not set before in the account saving goal", func(t *testing.T) {
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
			}, eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: "test-account",
				Version:    2,
				Event: eventually.Event{
					Payload: account.SavingGoalWasChanged{
						SavingGoal: saving.Goal{
							Amount:     500,
							Thresholds: []float64{0.25, 0.5},
						},
					},
				},
			}).
			When(eventually.Command{
				Payload: account.SetNewThreshold{
					AccountID: "test-account",
					Value:     0.75,
				},
			}).
			Then(eventstore.Event{
				StreamType: account.Type.Name(),
				StreamName: "test-account",
				Version:    3,
				Event: eventually.Event{
					Payload: account.ThresholdWasSet{
						Threshold: 0.75,
					},
				},
			}).
			Using(t, account.Type, func(r *aggregate.Repository) command.Handler {
				return account.SetNewThresholdCommandHandler{Repository: r}
			})
	})
}
