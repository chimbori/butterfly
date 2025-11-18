package main

import (
	"testing"

	"go.chimbori.app/butterfly/conf"
)

func TestValidateUrl_RejectsUnauthorizedDomains(t *testing.T) {
	// Setup: Configure authorized domains
	conf.Config = &conf.AppConfig{}
	conf.Config.LinkPreview.Domains = []string{"chimbori.com", "manas.tungare.name"}

	tests := []struct {
		name        string
		url         string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Authorized domain - chimbori.com",
			url:         "https://chimbori.com/page",
			shouldError: false,
		},
		{
			name:        "Authorized domain - subdomain of chimbori.com",
			url:         "https://apps.chimbori.com/page",
			shouldError: false,
		},
		{
			name:        "Authorized domain - manas.tungare.name",
			url:         "https://manas.tungare.name/article",
			shouldError: false,
		},
		{
			name:        "Unauthorized domain - google.com",
			url:         "https://google.com",
			shouldError: true,
			errorMsg:    "domain google.com not authorized",
		},
		{
			name:        "Unauthorized non-SSL domain - example.com",
			url:         "http://example.com/test",
			shouldError: true,
			errorMsg:    "domain example.com not authorized",
		},
		{
			name:        "Unauthorized domain - malicious.chimbori.com.evil.com",
			url:         "https://malicious.chimbori.com.evil.com",
			shouldError: true,
			errorMsg:    "domain malicious.chimbori.com.evil.com not authorized",
		},
		{
			name:        "Unauthorized domain - chimboricom (no dot)",
			url:         "https://chimboricom.attacker.com",
			shouldError: true,
			errorMsg:    "domain chimboricom.attacker.com not authorized",
		},
		{
			name:        "Unauthorized non-SSL subdomain",
			url:         "http://unauthorized.example.com",
			shouldError: true,
			errorMsg:    "domain unauthorized.example.com not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateUrl(tt.url)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for URL %s, but got none", tt.url)
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', but got '%s'", tt.errorMsg, err.Error())
				}
				if result != "" {
					t.Errorf("Expected empty result for unauthorized domain, but got %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for URL %s, but got: %s", tt.url, err.Error())
				}
				if result == "" {
					t.Errorf("Expected non-empty result for authorized domain")
				}
			}
		})
	}
}

func TestValidateUrl_EmptyUrl(t *testing.T) {
	_, err := validateUrl("")
	if err == nil {
		t.Error("Expected error for empty URL")
	}
	if err.Error() != "missing url" {
		t.Errorf("Expected 'missing url' error, got: %s", err.Error())
	}
}

func TestValidateUrl_InvalidUrl(t *testing.T) {
	conf.Config = &conf.AppConfig{}
	conf.Config.LinkPreview.Domains = []string{"chimbori.com"}

	_, err := validateUrl("ht!tp://invalid url with spaces")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestValidateUrl_AddsHttpsPrefix(t *testing.T) {
	conf.Config = &conf.AppConfig{}
	conf.Config.LinkPreview.Domains = []string{"chimbori.com"}

	result, err := validateUrl("chimbori.com/page")
	if err != nil {
		t.Errorf("Expected no error, but got: %s", err.Error())
	}
	if result != "https://chimbori.com/page" {
		t.Errorf("Expected https:// prefix to be added, got: %s", result)
	}
}
