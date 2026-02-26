package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 30 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

type Server struct {
	http *http.Server
	log  zerolog.Logger
}

func New(addr string, handler http.Handler, log zerolog.Logger) *Server {
	return &Server{
		log: log,
		http: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		s.log.Info().Str("addr", s.http.Addr).Msg("server started")
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		s.log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	s.log.Info().Msg("server stopped gracefully")
	return nil
}
