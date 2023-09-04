package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Set the rate limit and burst size
	rateLimit := 10
	rateBurst := 5

	// Create a new rate limiter
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)

	// Create a new test server with the rate limit middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.Header().Set("Retry-After", strconv.FormatInt(time.Now().Add(limiter.Reserve().Delay()).Unix(), 10))
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Vary", "User-Agent")
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, "OK")
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Send a large number of requests to the server
	client := server.Client()
	for i := 0; i < 100; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}
		defer resp.Body.Close()

		// Check if the response is "Too many requests" for the expected requests
		if i >= rateBurst && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, resp.StatusCode)
		} else if i < rateBurst && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	}
}
