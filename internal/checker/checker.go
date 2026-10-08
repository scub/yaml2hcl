// internal/check/checker.go
package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Endpoint struct {
	Name		string
	URL			string
	Method		string
	ExpectCode	int
	ExpectBody	string
}

type Result struct {
	Endpoint 	Endpoint
	Status		int
	Duration	time.Duration
	Body		string
	Err			error
	Healthy		bool
}

type Checker struct {
	client 		*http.Client
	verbose 	bool
}

func New(timeout time.Duration, verbose bool) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
		},
		verbose: verbose,
	}
}

func (c *Checker) CheckAll(ctx context.Context, endpoints []Endpoint) []Result {
	results := make([]Result, len(endpoints))
	var wg sync.WaitGroup

	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()
			results[idx] = c.check(ctx, endpoint)
		}(i, ep)
	}

	wg.Wait()
	return results
}

func (c *Checker) check(ctx context.Context, ep Endpoint) Result {
	result := Result{Endpoint: ep}

	req, err := http.NewRequestWithContext(ctx, ep.Method, ep.URL, nil)
	if err != nil {
		result.Err = fmt.Errorf("creating request: %w", err)
		return result
	}

	req.Header.Set("User-Agent", "healthcheck-cli/1.0")

	start := time.Now()
	resp, err := c.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Err = fmt.Errorf("request failed: %w", err)
		return result
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode

	// Read body (limit to 1MB to avoid memory issues)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		result.Err = fmt.Errorf("reading body: %w", err)
		return result
	}
	result.Body = string(bodyBytes)

	// Evaluate health
	result.Healthy = result.Status == ep.ExpectCode
	if ep.ExpectBody != "" && !strings.Contains(result.Body, ep.ExpectBody) {
		result.Healthy = false
	}

	return result
}

func (r *Result) Summary() string {
	icon := "OK"
	if !r.Healthy {
		icon = "FAIL"
	}
	if r.Err != nil {
		return fmt.Sprintf("[%s] %-25s error: %v", "ERR ", r.Endpoint.Name, r.Err)
	}
	return fmt.Sprintf("[%s] %-25s status=%d time=%dms",
		icon, r.Endpoint.Name, r.Status, r.Duration.Milliseconds())
}