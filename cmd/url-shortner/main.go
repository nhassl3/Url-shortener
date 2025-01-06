package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"urlshortner.com/m/internal/config"
	"urlshortner.com/m/internal/http-server/handlers/redirect"
	"urlshortner.com/m/internal/http-server/handlers/url/delete"
	"urlshortner.com/m/internal/http-server/handlers/url/save"
	"urlshortner.com/m/internal/http-server/middleware/logger"
	"urlshortner.com/m/internal/lib/logger/handlers/slogpretty"
	"urlshortner.com/m/internal/lib/logger/sl"
	"urlshortner.com/m/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config  cleanenv
	cfg := config.MustLoad()

	// init logger  slog
	log := setupLogger(cfg.Env)

	//log.With(slog.String("env", cfg.Env))
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init storage sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.ErrLog(err))
		os.Exit(1) // application failed with error (code - 1)
	}

	_ = storage

	// init router  chi, "chi render"
	router := chi.NewRouter()

	// middleware:
	router.Use(middleware.RequestID)
	//router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortner", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/delete/{alias}", delete.New(log, storage))
		r.Delete("/delete/", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))
	router.Get("/", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	// create object server and run him
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start server")
	}

	log.Error("SERVER STOPPED") // if some panic above
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
