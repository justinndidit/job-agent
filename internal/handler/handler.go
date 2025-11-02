package handler

import (
	"encoding/json"
	"net/http"

	"github.com/justinndidit/job-agent/internal/agent"
	"github.com/justinndidit/job-agent/internal/scraper"
	"github.com/rs/zerolog"
)

type Handler struct {
	executor *agent.AgentExecutor
	logger   *zerolog.Logger
}

func NewHandler(executor *agent.AgentExecutor, logger *zerolog.Logger) *Handler {
	return &Handler{executor: executor, logger: logger}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "job-agent",
	})
}

func (h *Handler) AgentCard(w http.ResponseWriter, r *http.Request) {
	card := map[string]interface{}{
		"name":         "Job Search Agent",
		"version":      "1.0.0",
		"description":  "AI-powered job search agent",
		"capabilities": []string{"Natural language job search"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

type JobSearchRequest struct {
	TaskID    string `json:"task_id"`
	RequestID string `json:"request_id"`
	Query     string `json:"query"`
}

type JobSearchResponse struct {
	Success bool                 `json:"success"`
	Count   int                  `json:"count"`
	Jobs    []scraper.JobPosting `json:"jobs"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (h *Handler) SearchJobs(w http.ResponseWriter, r *http.Request) {
	var req JobSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query required", http.StatusBadRequest)
		return
	}

	jobs, err := h.executor.SearchJobTool(r.Context(), req.Query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JobSearchResponse{
		Success: true,
		Count:   len(jobs),
		Jobs:    jobs,
	})
}
