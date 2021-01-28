package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eventually-rs/saving-goals-go/internal/domain/interval"

	"github.com/eventually-rs/eventually-go"
	"github.com/eventually-rs/eventually-go/eventstore"
	"github.com/go-chi/chi"
)

func forceMonthCreation(monthStore eventstore.Typed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		year, err := strconv.Atoi(chi.URLParam(r, "year"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		month, err := strconv.Atoi(chi.URLParam(r, "month"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = monthStore.
			Instance("month").
			Append(ctx, -1, eventually.Event{
				Payload: interval.MonthStarted{
					Month: interval.Month{
						Year:  year,
						Month: time.Month(month),
					},
				},
			})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
