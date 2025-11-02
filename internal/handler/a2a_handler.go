package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/justinndidit/job-agent/internal/agent"
	"github.com/justinndidit/job-agent/internal/scraper"
	"github.com/rs/zerolog"
)

type A2AHandler struct {
	executor    *agent.AgentExecutor
	logger      *zerolog.Logger
	telexAPIKey string
}

func NewA2AHandler(executor *agent.AgentExecutor, logger *zerolog.Logger, apiKey string) *A2AHandler {
	return &A2AHandler{
		executor:    executor,
		logger:      logger,
		telexAPIKey: apiKey,
	}
}

// A2A Protocol Types (JSON-RPC 2.0 with proper A2A structures)
type A2ARequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  A2AParams   `json:"params"`
	ID      interface{} `json:"id"`
}

type A2AParams struct {
	Message Message `json:"message"`
}

type Message struct {
	Role             string         `json:"role"`
	Parts            []Part         `json:"parts"`
	MessageID        string         `json:"messageId"`
	Kind             string         `json:"kind"`
	TaskID           string         `json:"taskId,omitempty"`
	ContextID        string         `json:"contextId,omitempty"`
	Extensions       []string       `json:"extensions,omitempty"`
	ReferenceTaskIDs []string       `json:"referenceTaskIds,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type Part struct {
	Kind string `json:"kind"` // "text", "file", or "data"
	Text string `json:"text,omitempty"`
}

type A2AResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  *Message    `json:"result,omitempty"`
	Error   *A2AError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type A2AError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// AgentCard returns the A2A agent card (GET /.well-known/agent.json)
func (h *A2AHandler) AgentCard(w http.ResponseWriter, r *http.Request) {
	card := map[string]interface{}{
		"name":        "Job Search Agent",
		"description": "AI-powered job search agent that finds relevant job postings using natural language queries",
		"url":         "https://involved-sheree-surgee-fcce11ee.koyeb.app", // Replace with your actual base URL
		"version":     "1.0.0",
		"provider": map[string]string{
			"organization": "justinndidit.org",
			"url":          "https://involved-sheree-surgee-fcce11ee.koyeb.app",
		},
		"capabilities": map[string]bool{
			"streaming":              false,
			"pushNotifications":      false,
			"stateTransitionHistory": false,
		},
		"defaultInputModes":  []string{"text/plain"},
		"defaultOutputModes": []string{"text/plain", "application/json"},
		"skills": []map[string]interface{}{
			{
				"id":          "job_search",
				"name":        "Job Search",
				"description": "Search for jobs by title and location using natural language",
				"inputModes":  []string{"text/plain"},
				"outputModes": []string{"text/plain"},
				"examples": []map[string]interface{}{
					{
						"input": map[string]interface{}{
							"parts": []map[string]string{
								{"text": "Find software engineer jobs in New York", "contentType": "text/plain"},
							},
						},
						"output": map[string]interface{}{
							"parts": []map[string]string{
								{"text": "Found 5 job opportunities...", "contentType": "text/plain"},
							},
						},
					},
				},
			},
		},
		"supportsAuthenticatedExtendedCard": false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// HandleA2A processes all A2A protocol requests (POST /)
func (h *A2AHandler) HandleA2A(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	apiKey := r.Header.Get("X-AGENT-API-KEY")
	if apiKey == "" {
		h.logger.Warn().Msg("Missing API key")
		h.sendError(w, nil, -32600, "Missing X-AGENT-API-KEY header")
		return
	}

	if h.telexAPIKey != "" && apiKey != h.telexAPIKey {
		h.logger.Warn().Str("key", maskKey(apiKey)).Msg("Invalid API key")
		h.sendError(w, nil, -32600, "Invalid API key")
		return
	}

	// Parse request
	var req A2ARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		h.sendError(w, nil, -32700, "Parse error")
		return
	}

	h.logger.Info().
		Str("method", req.Method).
		Str("role", req.Params.Message.Role).
		Msg("A2A request received")

	// Route based on method
	switch req.Method {
	case "message/send":
		h.handleMessageSend(w, r, &req)
	case "task/subscribe":
		h.sendError(w, req.ID, -32601, "Method not supported: task/subscribe")
	default:
		h.sendError(w, req.ID, -32601, "Method not found: "+req.Method)
	}
}

func (h *A2AHandler) handleMessageSend(w http.ResponseWriter, r *http.Request, req *A2ARequest) {
	// Extract text from parts
	var userQuery string
	for _, part := range req.Params.Message.Parts {
		if part.Kind == "text" && part.Text != "" {
			userQuery = part.Text
			break
		}
	}

	if userQuery == "" {
		h.sendError(w, req.ID, -32602, "No text content in message")
		return
	}

	// Execute search
	jobs, err := h.executor.SearchJobTool(r.Context(), userQuery)
	if err != nil {
		h.logger.Error().Err(err).Msg("Search failed")
		h.sendError(w, req.ID, -32603, "Search failed: "+err.Error())
		return
	}

	// Format response as A2A Message
	responseText := h.formatJobs(jobs)

	responseMessage := Message{
		Role:      "agent",
		Parts:     []Part{{Kind: "text", Text: responseText}},
		MessageID: generateMessageID(),
		Kind:      "message",
		Metadata: map[string]any{
			"jobCount":  len(jobs),
			"timestamp": time.Now().Unix(),
		},
	}

	// If the incoming message had a taskId or contextId, include them
	if req.Params.Message.TaskID != "" {
		responseMessage.TaskID = req.Params.Message.TaskID
	}
	if req.Params.Message.ContextID != "" {
		responseMessage.ContextID = req.Params.Message.ContextID
	}

	response := A2AResponse{
		JSONRPC: "2.0",
		Result:  &responseMessage,
		ID:      req.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *A2AHandler) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := A2AResponse{
		JSONRPC: "2.0",
		Error:   &A2AError{Code: code, Message: message},
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC always returns 200
	json.NewEncoder(w).Encode(response)
}

func (h *A2AHandler) formatJobs(jobs []scraper.JobPosting) string {
	if len(jobs) == 0 {
		return "No jobs found matching your criteria. Try different search terms or a broader location."
	}

	response := fmt.Sprintf("✨ Found %d job opportunities:\n\n", len(jobs))
	limit := 5
	if len(jobs) < limit {
		limit = len(jobs)
	}

	for i := 0; i < limit; i++ {
		job := jobs[i]
		location := "Remote"
		if len(job.JobLocation) > 0 {
			location = job.JobLocation[0]
		}

		response += fmt.Sprintf("%d. **%s** at %s\n", i+1, job.Title, job.Organization)
		response += fmt.Sprintf("   📍 %s", location)
		if job.Remote {
			response += " (Remote Available)"
		}
		response += fmt.Sprintf("\n   🔗 %s\n\n", job.SourceUrl)
	}

	if len(jobs) > limit {
		response += fmt.Sprintf("... and %d more jobs available!\n", len(jobs)-limit)
		response += "Try refining your search for more specific results."
	}

	return response
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
