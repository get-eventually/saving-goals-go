package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/aggregate"
	"github.com/eventually-rs/eventually-go/command"
	"github.com/eventually-rs/saving-goals-go/internal/domain/account"
	"github.com/eventually-rs/saving-goals-go/internal/domain/saving"
	"github.com/go-chi/chi"
)

func changeAccountSavingGoalHandler(commandBus command.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accountID := chi.URLParam(r, "accountId")

		var savingGoal saving.Goal
		if err := json.NewDecoder(r.Body).Decode(&savingGoal); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := commandBus.Dispatch(ctx, eventually.Command{
			Payload: account.ChangeSavingGoal{
				AccountID:  aggregate.StringID(accountID),
				SavingGoal: savingGoal,
			},
		})

		if errors.Is(err, account.ErrAtLeastOneThreshold) || errors.Is(err, account.ErrGoalIsZero) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

type SetNewThresholdRequest struct {
	Threshold float64 `json:"threshold"`
}

func setNewAccountSavingGoalThresholdHandler(commandBus command.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accountID := chi.URLParam(r, "accountId")

		var request SetNewThresholdRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := commandBus.Dispatch(ctx, eventually.Command{
			Payload: account.SetNewThreshold{
				AccountID: aggregate.StringID(accountID),
				Value:     request.Threshold,
			},
		})

		if errors.Is(err, account.ErrNoSavingGoal) || errors.Is(err, account.ErrThresholdAlreadyExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
