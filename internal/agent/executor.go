package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/justinndidit/job-agent/internal/scraper"
	"github.com/justinndidit/job-agent/internal/util"
	"github.com/rs/zerolog"
)

type AgentExecutor struct {
	scraper     *scraper.JobScraper
	geminiAgent *GeminiAgent
	logger      *zerolog.Logger
}

func NewExecutor(scraper *scraper.JobScraper, gemini *GeminiAgent, log *zerolog.Logger) *AgentExecutor {
	return &AgentExecutor{
		scraper:     scraper,
		geminiAgent: gemini,
		logger:      log,
	}
}

func (e *AgentExecutor) SearchJobTool(ctx context.Context, userQuery string) ([]scraper.JobPosting, error) {
	e.logger.Info().Str("query", userQuery).Msg("Processing job search")

	processedMessage, err := e.geminiAgent.ProcessQuery(ctx, userQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to process query: %w", err)
	}

	if strings.TrimSpace(processedMessage) == "invalid" {
		return nil, fmt.Errorf("query does not contain valid job search information")
	}

	query := util.ParseMessage(processedMessage)
	if query.Title == "" && query.Location == "" {
		return nil, fmt.Errorf("could not extract job information")
	}

	e.logger.Info().Str("title", query.Title).Str("location", query.Location).Msg("Parsed query")

	jobs, err := e.scraper.QueryJobs(ctx, &query)
	if err != nil {
		return nil, fmt.Errorf("failed to search jobs: %w", err)
	}

	e.logger.Info().Int("count", len(jobs)).Msg("Retrieved jobs")
	return jobs, nil
}
