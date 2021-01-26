package http

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

func NewRouter(commandBus command.Dispatcher, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(zapchi.UseLogger(logger)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, bytes.NewBufferString("{\"message\": \"Hello world!\"}\n"))
	})

	r.Route("/accounts/{id}", func(r chi.Router) {
		r.Post("/change-saving-goal", nil)
		r.Post("/set-new-threshold", nil)
	})

	return r
}
