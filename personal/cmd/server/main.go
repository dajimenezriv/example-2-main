package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chiapi "github.com/dajimenezriv/personal/internal/api"
	"github.com/dajimenezriv/personal/internal/db"
	"github.com/dajimenezriv/personal/internal/logger"
	"github.com/dajimenezriv/personal/internal/rate"
	"github.com/dajimenezriv/personal/internal/web"
)

const (
	serviceName  = "server"
	port         = ":8000"
	frontendPort = ":8080"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Setup()
	logger.With(ctx, "service_name", serviceName)

	l := rate.NewLimiter(ctx, 10, 1)
	m := chiapi.NewMiddleware(l)
	a := chiapi.NewLinksAPI(db.NewLinksStore(), m)

	srv := &http.Server{
		Addr:              port,
		Handler:           a.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.InfoContext(ctx, "start server", "port", port)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "listen and serve", "err", err)
			os.Exit(1)
		}
	}()

	frontend, err := web.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		slog.InfoContext(ctx, "listen frontend", "port", frontendPort)
		if err := http.ListenAndServe(frontendPort, frontend.Routes()); !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "listen frontend", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.InfoContext(ctx, "shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}
