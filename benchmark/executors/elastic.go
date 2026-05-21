package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type ElasticExecutor struct {
	url string
}

func NewElasticExecutor(url string) *ElasticExecutor {
	return &ElasticExecutor{url: url}
}

func (e *ElasticExecutor) RunQuery(ctx context.Context, query string) (time.Duration, error) {

	start := time.Now()

	req := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		e.url+"/orders/_search",
		bytes.NewBuffer(body),
	)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return time.Since(start), nil
}
