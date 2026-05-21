package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ElasticExecutor struct {
	url    string
	client *http.Client
}

func NewElasticExecutor(url string) *ElasticExecutor {
	return &ElasticExecutor{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *ElasticExecutor) RunQuery(ctx context.Context, query string) (time.Duration, error) {
	start := time.Now()

	// Determine endpoint and method based on query content
	endpoint := e.url + "/orders/_search"
	
	var bodyReader *bytes.Reader
	
	// Check if query is a full URL path (starts with GET/POST)
	if strings.HasPrefix(query, "GET ") || strings.HasPrefix(query, "POST ") {
		// Parse simple DSL format: "GET /index/_search {body}"
		parts := strings.SplitN(query, " ", 2)
		if len(parts) == 2 {
			bodyStr := parts[1]
			// Extract path from body if present
			if idx := strings.Index(bodyStr, "{"); idx > 0 {
				path := strings.TrimSpace(bodyStr[:idx])
				bodyStr = strings.TrimSpace(bodyStr[idx:])
				endpoint = e.url + path
			}
			bodyReader = bytes.NewReader([]byte(bodyStr))
		}
	} else {
		// Assume query is JSON body
		bodyReader = bytes.NewReader([]byte(query))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bodyReader)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read response body to ensure complete request
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("elastic query failed: %s", resp.Status)
	}

	return time.Since(start), nil
}

// RunBulkQuery executes a bulk query (for indexing)
func (e *ElasticExecutor) RunBulkQuery(ctx context.Context, body []byte) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "POST", e.url+"/_bulk", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// RunAggregation runs an aggregation query and returns results
func (e *ElasticExecutor) RunAggregation(ctx context.Context, aggQuery string) (map[string]interface{}, time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "POST", e.url+"/orders/_search", bytes.NewReader([]byte(aggQuery)))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	return result, time.Since(start), nil
}
