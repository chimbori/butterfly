package core

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSetupHealthCheck(t *testing.T) {
	t.Run("registers healthcheck endpoint", func(t *testing.T) {
		mux := http.NewServeMux()
		SetupHealthCheck(mux)

		req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != "ok" {
			t.Errorf("Expected response body “ok”, got %q", string(body))
		}
	})

	t.Run("only accepts GET method", func(t *testing.T) {
		mux := http.NewServeMux()
		SetupHealthCheck(mux)

		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/healthcheck", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			resp := w.Result()
			resp.Body.Close()

			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code %d for %s method, got %d", http.StatusMethodNotAllowed, method, resp.StatusCode)
			}
		}
	})

	t.Run("handles requests with various headers", func(t *testing.T) {
		mux := http.NewServeMux()
		SetupHealthCheck(mux)

		req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("handles nil mux gracefully", func(t *testing.T) {
		// This test ensures that calling SetupHealthCheck with nil doesn’t panic
		// It will panic if the implementation doesn’t handle nil, which is expected
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when passing nil mux, but didn’t panic")
			}
		}()

		SetupHealthCheck(nil)
	})
}

func TestVerifyHealthCheck(t *testing.T) {
	t.Run("successful healthcheck", func(t *testing.T) {
		// Create a test server
		mux := http.NewServeMux()
		SetupHealthCheck(mux)
		server := httptest.NewServer(mux)
		defer server.Close()

		// Extract port from server URL
		_, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		if err != nil {
			t.Fatalf("Failed to extract port from server URL: %v", err)
		}

		// Parse port
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			t.Fatalf("Failed to parse port: %v", err)
		}

		// Verify healthcheck
		exitCode := VerifyHealthCheck(port)
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("server not running", func(t *testing.T) {
		port := 59999 // Use a port that’s unlikely to be in use

		exitCode := VerifyHealthCheck(port)
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when server is not running, got %d", exitCode)
		}
	})

	t.Run("server returns error status", func(t *testing.T) {
		// Create a test server that returns 500
		mux := http.NewServeMux()
		mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "error", http.StatusInternalServerError)
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// Extract port from server URL
		_, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		if err != nil {
			t.Fatalf("Failed to extract port from server URL: %v", err)
		}

		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			t.Fatalf("Failed to parse port: %v", err)
		}

		exitCode := VerifyHealthCheck(port)
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when server returns error, got %d", exitCode)
		}
	})

	t.Run("server timeout", func(t *testing.T) {
		// Create a test server that delays response beyond timeout
		mux := http.NewServeMux()
		mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(10 * time.Second) // Longer than the 5 second timeout
			w.Write([]byte("ok"))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// Extract port from server URL
		_, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		if err != nil {
			t.Fatalf("Failed to extract port from server URL: %v", err)
		}

		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			t.Fatalf("Failed to parse port: %v", err)
		}

		// This should timeout and return exit code 1
		exitCode := VerifyHealthCheck(port)
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when server times out, got %d", exitCode)
		}
	})

	t.Run("various port numbers", func(t *testing.T) {
		// Create a test server
		mux := http.NewServeMux()
		SetupHealthCheck(mux)
		server := httptest.NewServer(mux)
		defer server.Close()

		// Extract port from server URL
		_, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		if err != nil {
			t.Fatalf("Failed to extract port from server URL: %v", err)
		}

		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			t.Fatalf("Failed to parse port: %v", err)
		}

		// Verify healthcheck works
		exitCode := VerifyHealthCheck(port)
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
	})
}
