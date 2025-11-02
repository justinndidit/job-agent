package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/justinndidit/job-agent/internal/config"
	"github.com/rs/zerolog"
)

type JobScraper struct {
	logger *zerolog.Logger
	config config.JobScraperConfig
	client *http.Client
}

type JobQuery struct {
	Title    string `json:"title_filter"`
	Location string `json:"location_filter"`
}

type JobPosting struct {
	Title            string   `json:"title"`
	OrganizationUrl  string   `json:"organization_url"`
	DatePosted       string   `json:"date_posted"`
	DateValidThrough string   `json:"date_validthrough"`
	Organization     string   `json:"organization"`
	SourceUrl        string   `json:"url"`
	EmploymentType   []string `json:"employment_type"`
	JobLocation      []string `json:"locations_derived"`
	TimeZone         []string `json:"timezones_derived"`
	Remote           bool     `json:"remote_derived"`
}

func NewJobScraper(cfg config.JobScraperConfig, log *zerolog.Logger) *JobScraper {
	return &JobScraper{
		logger: log,
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *JobScraper) QueryJobs(ctx context.Context, job *JobQuery) ([]JobPosting, error) {
	params := url.Values{}
	params.Add("limit", "10")
	params.Add("offset", "0")
	params.Add("description_type", "text")

	if job.Title != "" {
		params.Add("title_filter", fmt.Sprintf("\"%s\"", job.Title))
	}
	if job.Location != "" {
		params.Add("location_filter", fmt.Sprintf("\"%s\"", job.Location))
	}

	fullURL := fmt.Sprintf("%s?%s", s.config.RAPID_API_BASE_URL, params.Encode())
	s.logger.Info().Str("url", fullURL).Msg("Querying jobs API")

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "TelexJobAgent/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-rapidapi-host", s.config.RAPID_API_HOST)
	req.Header.Set("x-rapidapi-key", s.config.RAPID_API_KEY)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn().Int("status", resp.StatusCode).Str("body", string(body)).Msg("API error")
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var jobPostings []JobPosting
	if err := json.Unmarshal(body, &jobPostings); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return jobPostings, nil
}
