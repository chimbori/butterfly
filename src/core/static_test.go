package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupHealthz(t *testing.T) {
	mux := http.NewServeMux()
	SetupHealthz(mux)

	t.Run("GET /healthz returns ok", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		expectedBody := "ok"
		if w.Body.String() != expectedBody {
			t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /healthz not allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/healthz", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("healthz with X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		req.Header.Set("X-Forwarded-For", "1.1.1.1")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != "ok" {
			t.Errorf("Expected body %q, got %q", "ok", w.Body.String())
		}
	})
}

func TestServeWebManifest(t *testing.T) {
	tests := []struct {
		name       string
		appName    string
		url        string
		themeColor string
	}{
		{
			name:       "basic manifest",
			appName:    "Example App",
			url:        "https://example.com",
			themeColor: "#000000",
		},
		{
			name:       "app with URL path",
			appName:    "Example App",
			url:        "https://example.com/app",
			themeColor: "#FF5733",
		},
		{
			name:       "app with empty theme color",
			appName:    "Example App",
			url:        "/",
			themeColor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			ServeWebManifest(mux, tt.appName, tt.url, tt.themeColor)

			req := httptest.NewRequest("GET", "/app.webmanifest", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
			contentType := w.Header().Get("Content-Type")
			expectedContentType := "application/manifest+json"
			if contentType != expectedContentType {
				t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
			}

			// Parse JSON to verify it’s valid
			var manifest map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &manifest); err != nil {
				t.Fatalf("Failed to parse manifest JSON: %v\nBody: %s", err, w.Body.String())
			}

			// Verify mandatory fields
			if name, ok := manifest["name"].(string); !ok || name != tt.appName {
				t.Errorf("Expected name %q, got %v", tt.appName, manifest["name"])
			}
			if startURL, ok := manifest["start_url"].(string); !ok || startURL != tt.url {
				t.Errorf("Expected start_url %q, got %v", tt.url, manifest["start_url"])
			}
			if themeColor, ok := manifest["theme_color"].(string); !ok || themeColor != tt.themeColor {
				t.Errorf("Expected theme_color %q, got %v", tt.themeColor, manifest["theme_color"])
			}
			if display, ok := manifest["display"].(string); !ok || display != "standalone" {
				t.Errorf("Expected display %q, got %v", "standalone", manifest["display"])
			}
			icons, ok := manifest["icons"].([]interface{})
			if !ok {
				t.Fatal("Expected icons to be an array")
			}
			if len(icons) != 1 {
				t.Errorf("Expected 1 icon, got %d", len(icons))
			}
			if len(icons) > 0 {
				icon, ok := icons[0].(map[string]interface{})
				if !ok {
					t.Fatal("Expected icon to be an object")
				}
				if src, ok := icon["src"].(string); !ok || src != "/static/favicon.svg" {
					t.Errorf("Expected icon src %q, got %v", "/static/favicon.svg", icon["src"])
				}
				if iconType, ok := icon["type"].(string); !ok || iconType != "image/svg+xml" {
					t.Errorf("Expected icon type %q, got %v", "image/svg+xml", icon["type"])
				}
				if sizes, ok := icon["sizes"].(string); !ok || sizes != "144x144" {
					t.Errorf("Expected icon sizes %q, got %v", "144x144", icon["sizes"])
				}
			}
		})
	}

	t.Run("POST method not allowed", func(t *testing.T) {
		mux := http.NewServeMux()
		ServeWebManifest(mux, "Test App", "https://example.com", "#000000")

		req := httptest.NewRequest("POST", "/app.webmanifest", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}
