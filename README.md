# Job Search Agent 🤖

```bash
  An AI-powered job search agent built with Go that integrates with Telex.im using the A2A (Agent-to-Agent) protocol. The agent uses Google's Gemini AI to process natural language queries and return relevant job postings.
```

## 🌟 Features

```bash
    Natural Language Processing: Understands queries like "software engineer jobs in NYC" or "remote data scientist positions"
  A2A Protocol Compliant: Fully implements the JSON-RPC 2.0 based A2A protocol
  AI-Powered: Uses Google Gemini to intelligently parse job search queries
  Location Parsing: Automatically converts abbreviations (NY → New York, CA → California)
  Secure Authentication: API key-based authentication for Telex integration
  Job Aggregation: Scrapes and aggregates jobs from multiple sources
```

## 🏗️ Architecture

```bash

    ┌─────────────┐
    │   Telex.im  │
    │   Client    │
    └──────┬──────┘
          │ A2A Protocol (JSON-RPC 2.0)
          ▼
    ┌─────────────────────────────────────┐
    │      Job Search Agent (Go)          │
    │                                     │
    │  ┌──────────────┐  ┌─────────────┐  │
    │  │  A2A Handler │→ │   Gemini    │  │
    │  │  (JSON-RPC)  │  │  AI Agent   │  │
    │  └──────┬───────┘  └─────────────┘  │
    │         │                           │
    │         ▼                           │
    │  ┌──────────────┐                   │
    │  │ Job Scraper  │                   │
    │  └──────────────┘                   │
    └─────────────────────────────────────┘
          │
          ▼
    ┌─────────────────┐
    │  Job Boards API │
    └─────────────────

```

## 🚀 Quick Start
 ### Prerequisites

  Go 1.21 or higher
  Google Gemini API access (via Vertex AI or API key)
  A Telex.im account

  ### Installation

  #### Clone the repository

  ```bash

     git clone https://github.com/yourusername/job-agent.git
     cd job-agent

  ```
  #### Install dependencies
    ```bash

           go mod download

    ```

  #### Configure Environmental Variables

  ```bash

    cp .env.sample .env

  ```

  #### Run Agent

  ```bash

    go run cmd/main.go

  ```
## 📡 API Endpoints

  ### A2A Protocol Endpoints (Telex Integration)
  ```bash

    GET /.well-known/agent.json
    Returns the agent card for discovery (public endpoint).

    POST /
    Handles all JSON-RPC 2.0 method calls (authenticated).
    Supported Methods:

    message/send - Process job search queries

    GET /health
    Health check endpoint.
    POST /api/search
    Direct job search (backward compatibility).

  ```