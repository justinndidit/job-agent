package util

import (
	"strings"

	"github.com/justinndidit/job-agent/internal/scraper"
)

func ParseMessage(msg string) scraper.JobQuery {
	query := scraper.JobQuery{}
	msg = strings.TrimSpace(msg)

	if msg == "" {
		return query
	}

	fields := strings.Split(msg, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		parts := strings.SplitN(field, ":", 2)

		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			query.Title = value
		case "location":
			query.Location = value
		}
	}

	return query
}
