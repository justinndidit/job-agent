package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/justinndidit/job-agent/internal/config"
	"github.com/rs/zerolog"
)

type Server struct {
	Config     *config.Config
	Logger     *zerolog.Logger
	httpServer *http.Server
}

func New(cfg *config.Config, logger *zerolog.Logger) *Server {
	server := &Server{
		Config: cfg,
		Logger: logger,
	}
	return server
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(30) * time.Second,
		WriteTimeout: time.Duration(30) * time.Second,
		IdleTimeout:  time.Duration(60) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}

	s.Logger.Info().
		Str("port", s.Config.Port).
		Msg("starting server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	return nil
}
