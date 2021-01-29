package monthly

import (
	"fmt"
	"sort"

	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
)

// Type defines the MonthlySpending aggregate type.
var Type = aggregate.NewType("monthly-spending", func() aggregate.Root {
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

	id                     ID
	startingBalance        float64
	currentBalance         float64
	desiredBalance         float64
	spendingLimit          float64
	thresholds             []float64
	lastTriggeredThreshold float64
}

func (ms Spending) AggregateID() aggregate.ID { return ms.id }

type SpendingTrackingStarted struct {
	ID              ID
	StartingBalance float64
	DesiredBalance  float64
	Thresholds      []float64
}

type TransactionWasRecorded struct {
	Amount float64
}

type SpendingLimitWasUpdated struct {
	SpendingLimit float64
}

type ThresholdWasReached struct {
	Threshold float64
}

func (ms *Spending) Apply(event eventually.Event) error {
	switch evt := event.Payload.(type) {
	case SpendingTrackingStarted:
		ms.id = evt.ID
		ms.startingBalance = evt.StartingBalance
		ms.currentBalance = evt.StartingBalance
		ms.desiredBalance = evt.DesiredBalance
		ms.thresholds = evt.Thresholds

	case TransactionWasRecorded:
		ms.currentBalance += evt.Amount

	case SpendingLimitWasUpdated:
		ms.spendingLimit = evt.SpendingLimit

	case ThresholdWasReached:
		ms.lastTriggeredThreshold = evt.Threshold

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

func (s *Spending) RecordTransaction(amount float64) error {
	newBalance := s.currentBalance + amount

	err := aggregate.RecordThat(s, eventually.Event{
		Payload: TransactionWasRecorded{Amount: amount},
	})

	if err != nil {
		return fmt.Errorf("monthly.RecordTransaction: failed to record domain event: %w", err)
	}

	if amount > 0 {
		return s.updateSpendingLimit(newBalance)
	}

	return s.triggerThresholdOverstepIfAny(newBalance)
}

func (s *Spending) triggerThresholdOverstepIfAny(newBalance float64) error {
	currentPercentage := (newBalance - s.desiredBalance) / s.spendingLimit

	triggeredThresholds := make([]float64, 0, len(s.thresholds))
	for _, threshold := range s.thresholds {
		// We are only filtering for thresholds that fall between the last triggered one
		// and the current percentage of the spending limit already spent.
		//
		// This will give us back those thresholds that the current spending
		// percentage have been surpassed.
		if threshold <= s.lastTriggeredThreshold || currentPercentage < threshold {
			continue
		}

		triggeredThresholds = append(triggeredThresholds, threshold)
	}

	if len(triggeredThresholds) == 0 {
		// No triggered thresholds, no relevant Domain Events to register.
		return nil
	}

	// Let's take only the latest threshold surpassed, in case there are more than one,
	// which should be the highest one, after sorting.
	sort.Float64s(triggeredThresholds)
	triggeredThreshold := triggeredThresholds[len(triggeredThresholds)-1]

	err := aggregate.RecordThat(s, eventually.Event{
		Payload: ThresholdWasReached{Threshold: triggeredThreshold},
	})

	if err != nil {
		return fmt.Errorf("monthly.RecordTransaction.triggerThresholdOverstep: failed to record domain event: %w", err)
	}

	return nil
}

func (s *Spending) updateSpendingLimit(newBalance float64) error {
	err := aggregate.RecordThat(s, eventually.Event{
		Payload: SpendingLimitWasUpdated{SpendingLimit: newBalance - s.desiredBalance},
	})

	if err != nil {
		return fmt.Errorf("monthly.RecordTransaction.updateSpendingLimit: failed to record domain event: %w", err)
	}

	return nil
}
