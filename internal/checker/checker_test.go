// internal/checker/checker_test.go
package checker

import (
    "context"
    "net/http"
    "net/http/httptest"
	"sync/atomic"
    "testing"
    "time"
)

func TestCheck_HealthyEndpoint(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    c := New(5*time.Second, false)

    ep := Endpoint{
        Name:       "Test API",
        URL:        server.URL,
        Method:     "GET",
        ExpectCode: 200,
        ExpectBody: `"status":"ok"`,
    }

    result := c.check(context.Background(), ep)

    if result.Err != nil {
        t.Fatalf("unexpected error: %v", result.Err)
    }
    if !result.Healthy {
        t.Errorf("expected healthy, got unhealthy (status=%d)", result.Status)
    }
    if result.Duration == 0 {
        t.Error("expected non-zero duration")
    }
}

func TestCheck_UnhealthyStatusCode(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusServiceUnavailable)
    }))
    defer server.Close()

    c := New(5*time.Second, false)

    ep := Endpoint{
        Name:       "Failing API",
        URL:        server.URL,
        Method:     "GET",
        ExpectCode: 200,
    }

    result := c.check(context.Background(), ep)

    if result.Err != nil {
        t.Fatalf("unexpected error: %v", result.Err)
    }
    if result.Healthy {
        t.Error("expected unhealthy, got healthy")
    }
}

func TestCheck_BodyMismatch(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"degraded"}`))
    }))
    defer server.Close()

    c := New(5*time.Second, false)

    ep := Endpoint{
        Name:       "Degraded API",
        URL:        server.URL,
        Method:     "GET",
        ExpectCode: 200,
        ExpectBody: `"status":"ok"`,
    }

    result := c.check(context.Background(), ep)

    if result.Healthy {
        t.Error("expected unhealthy due to body mismatch")
    }
}

func TestCheck_Timeout(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    c := New(100*time.Millisecond, false)

    ep := Endpoint{
        Name:       "Slow API",
        URL:        server.URL,
        Method:     "GET",
        ExpectCode: 200,
    }

    result := c.check(context.Background(), ep)

    if result.Err == nil {
        t.Error("expected timeout error")
    }
}

func TestCheckAll_Concurrent(t *testing.T) {
	var callCount atomic.Int64
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        callCount.Add(1)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    c := New(5*time.Second, false)

    endpoints := []Endpoint{
        {Name: "EP1", URL: server.URL, Method: "GET", ExpectCode: 200},
        {Name: "EP2", URL: server.URL, Method: "GET", ExpectCode: 200},
        {Name: "EP3", URL: server.URL, Method: "GET", ExpectCode: 200},
    }

    results := c.CheckAll(context.Background(), endpoints)

    if len(results) != 3 {
        t.Fatalf("expected 3 results, got %d", len(results))
    }

    for i, r := range results {
        if !r.Healthy {
            t.Errorf("endpoint %d: expected healthy", i)
        }
    }
}

func TestResult_Summary(t *testing.T) {
    r := Result{
        Endpoint: Endpoint{Name: "Test"},
        Status:   200,
        Duration: 150 * time.Millisecond,
        Healthy:  true,
    }

    summary := r.Summary()
    if summary == "" {
        t.Error("expected non-empty summary")
    }
}