package monthly

import (
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
)

// Type defines the MonthlySpending aggregate type.
var Type = aggregate.NewType("montly-spending", func() aggregate.Root {
	return new(Spending)
})

// ID is the primary identifier type for a MonthlySpending Aggregate instance.
type ID struct {
	AccountID string
	Month     interval.Month
}

func (id ID) String() string {
	return fmt.Sprintf("account:%s:month:%s", id.AccountID, id.Month.String())
}

type Spending struct {
	aggregate.BaseRoot

	id              ID
	startingBalance float64
	currentBalance  float64
	desiredBalance  float64
	spendingLimit   float64
	thresholds      []float64
}

func (ms Spending) AggregateID() aggregate.ID { return ms.id }

type SpendingTrackingStarted struct {
	ID              ID
	StartingBalance float64
	DesiredBalance  float64
	Thresholds      []float64
}

func (ms *Spending) Apply(event eventually.Event) error {
	switch evt := event.Payload.(type) {
	case SpendingTrackingStarted:
		ms.id = evt.ID
		ms.startingBalance = evt.StartingBalance
		ms.desiredBalance = evt.DesiredBalance
		ms.thresholds = evt.Thresholds

	default:
		return fmt.Errorf("spending: unsupported event received")
	}

	return nil
}

func NewSpending(accountID string, month interval.Month, balance float64, goal saving.Goal) (*Spending, error) {
	var spending Spending

	err := aggregate.RecordThat(&spending, eventually.Event{
		Payload: SpendingTrackingStarted{
			ID: ID{
				AccountID: accountID,
				Month:     month,
			},
			StartingBalance: balance,
			DesiredBalance:  balance + goal.Amount,
			Thresholds:      goal.Thresholds,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("monthly.NewSpending: failed to record domain event: %w", err)
	}

	return &spending, nil
}
