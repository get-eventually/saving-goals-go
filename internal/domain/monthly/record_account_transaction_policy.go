package monthly

import (
	"context"
	"fmt"

	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"
	"go.uber.org/zap"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/eventually-rs/eventually-go/projection"
)

var _ projection.Applier = RecordTransactionPolicy{}

type RecordTransactionPolicy struct {
	CommandDispatcher command.Dispatcher
	Logger            *zap.Logger
}

func (rtp RecordTransactionPolicy) Apply(ctx context.Context, evt eventstore.Event) error {
	if event, ok := evt.Payload.(account.TransactionWasRecorded); ok {
		err := rtp.CommandDispatcher.Dispatch(ctx, eventually.Command{
			Payload: RecordTransaction{
				ID: ID{
					AccountID: evt.StreamName,
					Month:     interval.MonthFromTime(event.HappenedAt),
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
