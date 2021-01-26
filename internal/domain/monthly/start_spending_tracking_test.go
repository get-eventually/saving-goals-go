package monthly_test

import (
	"testing"
	"time"

	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"github.com/eventually-rs/saving-goals-go/internal/domain/monthly"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/scenario"
)

func TestStartSpendingTracking(t *testing.T) {
	t.Run("spending creation works when the spending was not already started", func(t *testing.T) {
		monthlySpendingID := monthly.ID{
			AccountID: "test-account",
			Month:     interval.MonthFromTime(time.Now()),
		}

		scenario.
			CommandHandler().
			When(eventually.Command{
				Payload: monthly.StartSpendingTracking{
					AccountID:       monthlySpendingID.AccountID,
					Month:           monthlySpendingID.Month,
					StartingBalance: 1000,
					SavingGoal: saving.Goal{
						Amount:     500,
						Thresholds: []float64{0.25, 0.5},
					},
				},
			}).
			Then(eventstore.Event{
				StreamType: monthly.Type.Name(),
				StreamName: monthlySpendingID.String(),
				Version:    1,
				Event: eventually.Event{
					Payload: monthly.SpendingTrackingStarted{
						ID:              monthlySpendingID,
						StartingBalance: 1000,
						DesiredBalance:  1500,
						Thresholds:      []float64{0.25, 0.5},
					},
				},
			}).
			Using(t, monthly.Type, func(r *aggregate.Repository) command.Handler {
				return monthly.StartSpendingTrackingCommandHandler{Repository: r}
			})
	})

	t.Run("spending creation fails if the spending was already started", func(t *testing.T) {
		monthlySpendingID := monthly.ID{
			AccountID: "test-account",
			Month:     interval.MonthFromTime(time.Now()),
		}

		scenario.
			CommandHandler().
			Given(eventstore.Event{
				StreamType: monthly.Type.Name(),
				StreamName: monthlySpendingID.String(),
				Version:    1,
				Event: eventually.Event{
					Payload: monthly.SpendingTrackingStarted{
						ID:              monthlySpendingID,
						StartingBalance: 1000,
						DesiredBalance:  1500,
						Thresholds:      []float64{0.25, 0.5},
					},
				},
			}).
			When(eventually.Command{
				Payload: monthly.StartSpendingTracking{
					AccountID:       monthlySpendingID.AccountID,
					Month:           monthlySpendingID.Month,
					StartingBalance: 1000,
					SavingGoal: saving.Goal{
						Amount:     500,
						Thresholds: []float64{0.25, 0.5},
					},
				},
			}).
			ThenFails().
			Using(t, monthly.Type, func(r *aggregate.Repository) command.Handler {
				return monthly.StartSpendingTrackingCommandHandler{Repository: r}
			})
	})
}
