package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"url-shortener/internal/common"
	"url-shortener/internal/stats"
)

func main() {
	log := common.NewLogger("stats-api")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dsn := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/urlshort?sslmode=disable")
	addr := env("ADDR", ":8081")

	pool, err := common.NewPool(ctx, dsn)
	if err != nil {
		log.Error("db init failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := stats.NewRepository(pool)
	svc := stats.NewService(repo)
	h := stats.NewHandler(svc)

	metrics := common.NewMetrics(nil, "stats")

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/stats/{code}", metrics.Observe(log, "/stats/{code}", h.Get))

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
