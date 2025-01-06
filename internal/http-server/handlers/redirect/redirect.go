package redirect

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

const opNew = "handlers.redirect.New"

//go:generate go run github.com/vektra/mockery/v3 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
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

		resURL, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("URL not found", slog.String("alias", alias))
				render.JSON(w, r, response.Error("URL not found"))
				return
			}
			log.Error("Failed to get url", sl.ErrLog(err))
			render.JSON(w, r, response.Error("Internal error"))
			return
		}

		log.Info("Got URL", slog.String("url", resURL))
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
