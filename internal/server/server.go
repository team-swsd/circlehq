package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// NewHTTPServer serve a new http rest server
func NewHTTPServer(handler http.Handler, addr string, sig []os.Signal, maxShutdownWait time.Duration, logger *slog.Logger) (err error) {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		s := make(chan os.Signal, 1)
		defer close(s)

		signal.Notify(s, sig...)
		defer signal.Stop(s)

		<-s

		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownWait)
		defer cancel()

		err = srv.Shutdown(ctx)
		logger.InfoContext(ctx, "Shutdown Server")
	}()

	logger.Info("start server", "address", addr)
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return err
}
