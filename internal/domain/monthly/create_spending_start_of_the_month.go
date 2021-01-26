package monthly

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/projection"
	"github.com/eventually-rs/eventually-go/query"
)

var _ projection.Applier = CreateSpendingStartOfTheMonthPolicy{}

type QueryDispatcher interface {
	Dispatch(context.Context, query.Query) (query.Answer, error)
}

type CreateSpendingStartOfTheMonthPolicy struct {
	CommandDispatcher command.Dispatcher
	QueryDispatcher   QueryDispatcher
}

func (csp CreateSpendingStartOfTheMonthPolicy) Apply(ctx context.Context, event eventstore.Event) error {
	if event, ok := event.Payload.(interval.MonthStarted); ok {
		return csp.handleMonthStarted(ctx, event.Month)
	}

	return nil
}

func (csp CreateSpendingStartOfTheMonthPolicy) handleMonthStarted(ctx context.Context, month interval.Month) error {
	answer, err := csp.QueryDispatcher.Dispatch(ctx, account.WithSavingGoalsQuery{})
	if err != nil {
		return fmt.Errorf("monthly.CreateSpendingStartOfTheMonthPolicy: failed to list accounts: %w", err)
	}

	accounts := answer.(account.WithSavingGoalsAnswer)
	for account := range accounts {
		if account.CurrentBalance < account.SavingGoal.Amount {
			continue
		}

		err := csp.CommandDispatcher.Dispatch(ctx, eventually.Command{
			Payload: StartSpendingTracking{
				Month:           month,
				AccountID:       account.AccountID,
				StartingBalance: account.CurrentBalance,
				SavingGoal:      account.SavingGoal,
			},
		})

		if err != nil {
			return fmt.Errorf("monthly.CreateSpendingStartOfTheMonthPolicy: failed to dispatch command: %w", err)
		}
	}

	return nil
}
