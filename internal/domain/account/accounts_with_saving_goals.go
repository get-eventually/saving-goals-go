package account

import (
	"context"
	"sync"

	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/projection"
	"github.com/eventually-rs/eventually-go/query"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"
)

var _ projection.Projection = &WithSavingGoalsProjection{}

// WithSavingGoalsQuery is the Domain Query used to fetch all the
// Accounts that have a Saving Goal set, alongisde their current Balance.
type WithSavingGoalsQuery struct {
	// BufferSize specifies how many entities the channel returned as
	// Domain Answer should buffer.
	BufferSize int
}

// WithSavingGoalsAnswer is the Domain Answer returned from an
// AccountsWithSavingGoalsQuery, which is a channel containing all the
// Accounts with a Saving Goal set and their current Balance.
type WithSavingGoalsAnswer <-chan WithSavingGoal

// WithSavingGoal represents a single Account projection,
// containing its current Balance and its Saving Goal amount.
type WithSavingGoal struct {
	AccountID      string
	CurrentBalance float64
	SavingGoal     saving.Goal
}

// WithSavingGoalsProjection listens to Account Domain Events to build
// a list of Accounts with their Saving Goals and latest Account Balance.
type WithSavingGoalsProjection struct {
	mx       sync.RWMutex
	accounts map[string]withSavingGoalEntry
}

type withSavingGoalEntry struct {
	balance    float64
	savingGoal *saving.Goal
}

// NewWithSavingGoalsProjection returns a new instance of WithSavingGoalsProjection type.
func NewWithSavingGoalsProjection() *WithSavingGoalsProjection {
	return &WithSavingGoalsProjection{
		accounts: make(map[string]withSavingGoalEntry),
	}
}

// QueryType binds the WithSavingGoalsQuery type to the projection.
func (*WithSavingGoalsProjection) QueryType() query.Query { return WithSavingGoalsQuery{} }

// Apply updates the state of the projection using the incoming event.
func (p *WithSavingGoalsProjection) Apply(ctx context.Context, event eventstore.Event) error {
	p.mx.Lock()
	defer p.mx.Unlock()

	switch evt := event.Payload.(type) {
	case WasCreated:
		p.accounts[evt.AccountID] = withSavingGoalEntry{}

	case SavingGoalWasChanged:
		entry := p.accounts[event.StreamName]
		entry.savingGoal = &evt.SavingGoal
		p.accounts[event.StreamName] = entry

	case ThresholdWasSet:
		entry := p.accounts[event.StreamName]
		entry.savingGoal.Thresholds = append(entry.savingGoal.Thresholds, evt.Threshold)
		p.accounts[event.StreamName] = entry

	case SavingGoalWasDisabled:
		entry := p.accounts[event.StreamName]
		entry.savingGoal = nil
		p.accounts[event.StreamName] = entry

	case TransactionWasRecorded:
		entry := p.accounts[event.StreamName]
		entry.balance += evt.Amount
		p.accounts[event.StreamName] = entry
	}

	return nil
}

// Handle returns a channel containing the list of all Accounts that have set
// a Saving Goal, if any.
//
// The channel created is buffered following the input provided in
// WithSavingGoalsQuery.
//
// The channel returned will close either when the result set is exhausted,
// or when the provided context is closed.
func (p *WithSavingGoalsProjection) Handle(ctx context.Context, query query.Query) (query.Answer, error) {
	q := query.(WithSavingGoalsQuery)
	ch := make(chan WithSavingGoal, q.BufferSize)

	go func() {
		p.mx.RLock()
		defer p.mx.RUnlock()
		defer close(ch)

		for id, entry := range p.accounts {
			if entry.savingGoal == nil {
				continue
			}

			item := WithSavingGoal{
				AccountID:      id,
				CurrentBalance: entry.balance,
				SavingGoal:     *entry.savingGoal,
			}

			select {
			case ch <- item:
				continue

			case <-ctx.Done():
				// If the writing operation on the channel is blocking and the context is canceled
				// in the meantime, drop efforts in populating the channel, as the operation is likely meaningless now.
				return
			}
		}
	}()

	return WithSavingGoalsAnswer(ch), nil
}
