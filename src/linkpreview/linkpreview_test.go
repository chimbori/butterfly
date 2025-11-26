package linkpreview

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.chimbori.app/butterfly/core"
	"go.chimbori.app/butterfly/db"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, *db.Queries) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Connect to the database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://chimbori:chimbori@localhost:5432/butterfly"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}

	// Clean up domains table before each test
	_, err = pool.Exec(context.Background(), "DELETE FROM domains")
	if err != nil {
		pool.Close()
		t.Fatalf("Unable to clean domains table: %v", err)
	}

	queries := db.New(pool)
	return pool, queries
}

func TestValidateUrl_RejectsUnauthorizedDomains(t *testing.T) {
	pool, queries := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// Setup: Add authorized domains to database
	_, err := queries.UpsertDomain(ctx, db.UpsertDomainParams{
		Domain:            "chimbori.com",
		IncludeSubdomains: core.Ptr(true),
		Authorized:        core.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Failed to insert test domain: %v", err)
	}

	_, err = queries.UpsertDomain(ctx, db.UpsertDomainParams{
		Domain:            "manas.tungare.name",
		IncludeSubdomains: core.Ptr(false),
		Authorized:        core.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Failed to insert test domain: %v", err)
	}

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
			result, err := validateUrl(ctx, queries, tt.url)

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
	pool, queries := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	_, err := validateUrl(ctx, queries, "")
	if err == nil {
		t.Error("Expected error for empty URL")
	}
	if err.Error() != "missing url" {
		t.Errorf("Expected 'missing url' error, got: %s", err.Error())
	}
}

func TestValidateUrl_InvalidUrl(t *testing.T) {
	pool, queries := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	_, err := validateUrl(ctx, queries, "ht!tp://invalid url with spaces")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestValidateUrl_AddsHttpsPrefix(t *testing.T) {
	pool, queries := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// Add authorized domain to database
	_, err := queries.UpsertDomain(ctx, db.UpsertDomainParams{
		Domain:            "chimbori.com",
		IncludeSubdomains: core.Ptr(true),
		Authorized:        core.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Failed to insert test domain: %v", err)
	}

	result, err := validateUrl(ctx, queries, "chimbori.com/page")
	if err != nil {
		t.Errorf("Expected no error, but got: %s", err.Error())
	}
	if result != "https://chimbori.com/page" {
		t.Errorf("Expected https:// prefix to be added, got: %s", result)
	}
}
