package httpapi

import (
	"bytes"
	"io"
	"net/http"

	"github.com/eventually-rs/saving-goals-go/pkg/zapchi"

	"github.com/eventually-rs/eventually-go/command"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// NewRouter returns a new instance of the HTTP API router.
func NewRouter(commandBus command.Dispatcher, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(zapchi.UseLogger(logger)))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, bytes.NewBufferString("{\"message\": \"Hello world!\"}\n"))
	})

	r.Route("/accounts/{accountId}", func(r chi.Router) {
		r.Post("/change-saving-goal", changeAccountSavingGoalHandler(commandBus))
		r.Post("/set-new-threshold", setNewAccountSavingGoalThresholdHandler(commandBus))
	})

	return r
}
