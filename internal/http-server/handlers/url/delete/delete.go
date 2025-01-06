package delete

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"urlshortner.com/m/internal/lib/api/response"
	"urlshortner.com/m/internal/lib/logger/sl"
	"urlshortner.com/m/internal/storage"
)

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v3 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	const opNew = "handlers.url.delete.New"

	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			slog.String("op", opNew),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			alias = r.URL.Query().Get("alias")
			if alias == "" {
				log.Info("Alias is empty")
				render.JSON(w, r, response.Error("Alias not found"))
				return
			}
		}

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasNoExists) {
				log.Info("Alias not found", slog.String("alias", alias))
				render.JSON(w, r, response.Error("Alias not found"))
				return
			}
			log.Error("Failed to delete url", sl.ErrLog(err))
			render.JSON(w, r, response.Error("Deleting error"))
			return
		}

		// TODO: insert code in other function for simple reading
		log.Info("URL was delete")
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
