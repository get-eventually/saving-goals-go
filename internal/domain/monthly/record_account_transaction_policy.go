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
)

var _ projection.Applier = RecordTransactionPolicy{}

type RecordTransactionPolicy struct {
	CommandDispatcher command.Dispatcher
}

func (rtp RecordTransactionPolicy) Apply(ctx context.Context, evt eventstore.Event) error {
	if event, ok := evt.Payload.(account.RecordTransaction); ok {
		err := rtp.CommandDispatcher.Dispatch(ctx, eventually.Command{
			Payload: RecordTransaction{
				ID: ID{
					AccountID: event.AccountID.String(),
					Month:     interval.Month{}, // TODO: fix this
				},
				Amount: event.Amount,
			},
		})

		if err != nil {
			return fmt.Errorf("monthly.RecordTransactionPolicy: failed to dispatch command: %w", err)
		}
	}

	return nil
}
