package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"github.com/justinndidit/job-agent/internal/agent"
	"github.com/justinndidit/job-agent/internal/config"
	"github.com/justinndidit/job-agent/internal/handler"
	"github.com/justinndidit/job-agent/internal/logger"
	"github.com/justinndidit/job-agent/internal/scraper"
)

func main() {
	// Setup logger
	log := logger.NewLoggerWithService("job-agent")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// if cfg.TelexAPIKey == "" {
	// 	log.Warn().Msg("TELEX_API_KEY not set - A2A auth disabled")
	// }

	// Initialize components
	jobScraper := scraper.NewJobScraper(cfg.JobScraper, &log)
	geminiAgent := agent.NewGeminiAgent(&log)
	executor := agent.NewExecutor(jobScraper, geminiAgent, &log)

	// Initialize handlers
	regularHandler := handler.NewHandler(executor, &log)
	a2aHandler := handler.NewA2AHandler(executor, &log) //, cfg.TelexAPIKey)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", regularHandler.HealthCheck)

	// ===== A2A Protocol Endpoints (Telex Standard) =====
	// These are the REQUIRED endpoints for A2A protocol compliance

	// 1. Agent Card Discovery - PUBLIC endpoint
	//    Telex calls this to discover your agent's capabilities
	//    Must be at: /.well-known/agent.json
	r.Get("/.well-known/agent.json", a2aHandler.AgentCard)

	// 2. RPC Endpoint - AUTHENTICATED endpoint
	//    All method calls (message/send, task/subscribe, etc.) go here
	//    Must be at: / (root)
	r.Post("/", a2aHandler.HandleA2A)

	// ===== Legacy API Routes (Optional - For Testing) =====
	// You can keep these for backward compatibility or testing
	r.Route("/api", func(r chi.Router) {
		r.Get("/agent-card", regularHandler.AgentCard)
		r.Post("/search", regularHandler.SearchJobs)
	})

	// Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start
	go func() {
		log.Info().
			Str("port", cfg.Port).
			// Bool("telex_auth", cfg.TelexAPIKey != "").
			Msg("Starting A2A-compliant Job Search Agent")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}

	geminiAgent.Close()
	log.Info().Msg("Server stopped gracefully")
}
