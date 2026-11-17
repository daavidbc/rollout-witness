package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessfulStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{201, true},
		{404, false},
		{503, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("HTTP_%d", test.statusCode), func(t *testing.T) {
			result := isSuccessfulStatus(test.statusCode)

			if result != test.expected {
				t.Errorf(
					"Status %d: expected %t, got %t",
					test.statusCode,
					test.expected,
					result,
				)
			}
		})
	}
}

func TestHTTPServer(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{404, false},
		{503, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("HTTP_%d", test.statusCode), func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(test.statusCode)
				}),
			)
			defer server.Close()

			response, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer response.Body.Close()

			result := isSuccessfulStatus(response.StatusCode)

			if result != test.expected {
				t.Errorf("Expected %t, got %t", test.expected, result)
			}
		})
	}
}

func TestConnectionError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	server.Close()

	response, err := http.Get(server.URL)

	if err == nil {
		if response != nil {
			response.Body.Close()
		}
		t.Fatal("Expected a connection error, but request succeeded")
	}
}
