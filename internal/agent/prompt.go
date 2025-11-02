package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"google.golang.org/genai"
)

type GeminiAgent struct {
	client     *genai.Client
	clientOnce sync.Once
	logger     *zerolog.Logger
}

func NewGeminiAgent(log *zerolog.Logger) *GeminiAgent {
	return &GeminiAgent{logger: log}
}

func (g *GeminiAgent) initClient(ctx context.Context) error {
	var initErr error
	g.clientOnce.Do(func() {
		client, err := genai.NewClient(ctx, nil)
		if err != nil {
			g.logger.Error().Err(err).Msg("Failed to create Gemini client")
			initErr = err
			return
		}
		g.client = client
		g.logger.Info().Msg("Gemini client initialized successfully")
	})
	return initErr
}

func (g *GeminiAgent) ProcessQuery(ctx context.Context, input string) (string, error) {
	if err := g.initClient(ctx); err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Extract job search information from this message.

				Message: %s

				Return format: "title: <job_title>, location: <location>"
				- Convert abbreviations (NY→New York, CA→California)
				- If no job info, return exactly "invalid"
				- Be flexible with informal language

				Examples:
				"software engineer job in SF" → "title: software engineer, location: San Francisco"
				"hello" → "invalid"`, input)

	result, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	response := strings.TrimSpace(result.Text())
	g.logger.Debug().Str("input", input).Str("output", response).Msg("Gemini processed query")
	return response, nil
}

func (g *GeminiAgent) Close() error {
	return nil
}
