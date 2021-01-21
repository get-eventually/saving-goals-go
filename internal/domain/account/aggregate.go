package account

import (
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
)

var (
	// ErrAtLeastOneThreshold is returned when no thresholds have been specified
	// when changing or setting an Account's Saving Goal.
	ErrAtLeastOneThreshold = fmt.Errorf("account.ChangeSavingGoal: at least one threshold should be specified")

	// ErrGoalIsZero is returned when changing or setting an Account's Saving Goal
	// using a zero-valued saving amount.
	ErrGoalIsZero = fmt.Errorf("account.ChangeSavingGoal: saving goal amount should be more than zero")

	// ErrNoSavingGoal is returned when trying to set a new Saving Goal threshold
	// for an Account, but the Account does not have a Saving Goal defined yet.
	ErrNoSavingGoal = fmt.Errorf("account.SetNewThreshold: saving goal was not set")

	// ErrThresholdAlreadyExists is returned when trying to set a new Saving Goal threshold
	// for an Account, but the specified threshold already exists.
	ErrThresholdAlreadyExists = fmt.Errorf("account.SetNewThreshold: threshold already exists")
)

// Type defines the Account aggregate type.
var Type = aggregate.NewType("account", func() aggregate.Root {
	return new(Account)
})

// Account is an Aggregate type that represents the Account in the
// Saving Goals bounded context, as it tracks both Account balance
// and Owner's defined Saving Goals.
type Account struct {
	aggregate.BaseRoot

	accountID  aggregate.StringID
	balance    float64
	savingGoal *saving.Goal
}

// AggregateID returns the accountId of the Account Aggregate.
func (a Account) AggregateID() aggregate.ID { return a.accountID }

// WasCreated is the Domain Event triggered by the Aggregate when
// a new Account instance is created.
type WasCreated struct {
	AccountID string
}

// SavingGoalWasChanged is the Domain Event triggered by the Aggregate
// when a new Saving Goal is chosed by the Account's Owner.
type SavingGoalWasChanged struct {
	SavingGoal saving.Goal
}

// ThresholdWasSet is the Domain Event triggered by the Aggregate
// when setting a new Threshold for the Account's Saving Goal.
type ThresholdWasSet struct {
	Threshold float64
}

// Apply applies the Domain Event received onto the Aggregate Root
// by mutating the Root's state accordingly.
func (a *Account) Apply(event eventually.Event) error {
	switch evt := event.Payload.(type) {
	case WasCreated:
		a.accountID = aggregate.StringID(evt.AccountID)
		a.balance = 0
		a.savingGoal = nil

	case SavingGoalWasChanged:
		a.savingGoal = &evt.SavingGoal

	case ThresholdWasSet:
		a.savingGoal.Thresholds = append(a.savingGoal.Thresholds, evt.Threshold)

	default:
		return fmt.Errorf("account: unsupported event received")
	}

	return nil
}

// Create creates a new Account instance, given the specified accountId.
//
// An error is returned if recording the Domain Event fails.
func Create(accountID string) (*Account, error) {
	var account Account

	err := aggregate.RecordThat(&account, eventually.Event{
		Payload: WasCreated{AccountID: accountID},
	})

	if err != nil {
		return nil, fmt.Errorf("account.Create: failed to record domain even: %w", err)
	}

	return &account, nil
}

// ChangeSavingGoal changes the Account's Saving Goal with the specified one.
//
// An error is returned if no thresholds have been specified in the
// new Saving goal, or if the Saving Goal target amount is zero.
func (a *Account) ChangeSavingGoal(goal saving.Goal) error {
	if len(goal.Thresholds) < 1 {
		return ErrAtLeastOneThreshold
	}

	if goal.Amount == 0 {
		return ErrGoalIsZero
	}

	err := aggregate.RecordThat(a, eventually.Event{
		Payload: SavingGoalWasChanged{SavingGoal: goal},
	})

	if err != nil {
		return fmt.Errorf("account.ChangeSavingGoal: failed to record domain even: %w", err)
	}

	return nil
}

// SetNewThreshold adds a new threshold to the Account's Saving Goal.
//
// ErrNoSavingGoal is returned if the Account has no Saving Goal set.
//
// ErrThresholdAlreadyExists is returned if the Account's Saving Goal already
// has the very same threshold set.
func (a *Account) SetNewThreshold(threshold float64) error {
	if a.savingGoal == nil {
		return ErrNoSavingGoal
	}

	for _, th := range a.savingGoal.Thresholds {
		if th == threshold {
			return ErrThresholdAlreadyExists
		}
	}

	err := aggregate.RecordThat(a, eventually.Event{
		Payload: ThresholdWasSet{Threshold: threshold},
	})

	if err != nil {
		return fmt.Errorf("account.SetNewThreshold: failed to record domain even: %w", err)
	}

	return nil
}
