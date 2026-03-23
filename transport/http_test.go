package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientGetSuccess(t *testing.T) {
	client := &HTTPClient{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodGet {
					t.Fatalf("expected GET, got %s", r.Method)
				}

				return responseWithStatus(http.StatusOK, "ok"), nil
			}),
		},
	}

	code, err := client.Get(context.Background(), "http://pulse.test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, code)
	}
}

func TestHTTPClientPostSuccess(t *testing.T) {
	client := &HTTPClient{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Fatalf("expected POST, got %s", r.Method)
				}

				payload, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("expected readable body, got %v", err)
				}

				if string(payload) != "pulse" {
					t.Fatalf("expected body %q, got %q", "pulse", string(payload))
				}

				return responseWithStatus(http.StatusCreated, ""), nil
			}),
		},
	}

	code, err := client.Post(context.Background(), "http://pulse.test", strings.NewReader("pulse"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, code)
	}
}

func TestHTTPClientReturnsErrorForFailingStatusCode(t *testing.T) {
	client := &HTTPClient{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return responseWithStatus(http.StatusInternalServerError, ""), nil
			}),
		},
	}

	code, err := client.Get(context.Background(), "http://pulse.test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, code)
	}
	var httpErr *HTTPStatusError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected *HTTPStatusError with 500, got %v", err)
	}
}

func TestHTTPClientWithAppliesHeaders(t *testing.T) {
	var gotPulse, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPulse = r.Header.Get("X-Pulse-Test")
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClientWith(HTTPClientConfig{
		Headers: map[string]string{
			"X-Pulse-Test": "hello",
			"Content-Type": "application/json",
		},
	})
	code, err := client.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, code)
	}
	if gotPulse != "hello" {
		t.Fatalf("X-Pulse-Test: want %q, got %q", "hello", gotPulse)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type: want %q, got %q", "application/json", gotCT)
	}
}

func TestHTTPClientWithUsesTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClientWith(HTTPClientConfig{Timeout: 40 * time.Millisecond})
	code, err := client.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if code != 0 {
		t.Fatalf("expected status code 0 before response, got %d", code)
	}
}

func TestHTTPClientRespectsContextCancellation(t *testing.T) {
	client := &HTTPClient{
		client: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				<-r.Context().Done()
				return nil, r.Context().Err()
			}),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	code, err := client.Get(ctx, "http://pulse.test")
	if code != 0 {
		t.Fatalf("expected status 0 before response, got %d", code)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected %v, got %v", context.Canceled, err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func responseWithStatus(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
