package slogdb

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/db"
)

// TestDBHandler tests that logs are written to the database as expected.
// This is an integration test that requires a running database.
func TestDBHandler(t *testing.T) {
	// Skip if not in integration test mode
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
	defer pool.Close()

	// Create the DB handler wrapping a tint handler
	tintHandler := tint.NewHandler(os.Stderr, &tint.Options{TimeFormat: "2006-01-02 15:04:05.000"})
	dbHandler := NewDBHandler(tintHandler, pool)

	// Create a logger with the DB handler
	logger := slog.New(dbHandler)

	// Log some messages at different levels
	logger.Info("This is an info message - should only go to console")
	logger.Warn("This is a warning - should only go to console")

	// Log an error - this should go to both console AND database
	logger.Error("This is an error message",
		"path", "/test/path",
		"method", "GET",
		"status", 500,
		"url", "https://example.com/test",
		"hostname", "example.com",
	)

	// Query the database to verify the error was logged
	queries := db.New(pool)
	recentLogs, err := queries.GetRecentLogs(context.Background(), 1)
	if err != nil {
		t.Fatalf("Failed to query recent logs: %v", err)
	}

	if len(recentLogs) == 0 {
		t.Fatal("Expected at least one log in the database")
	}

	lastLog := recentLogs[0]
	if lastLog.Message == nil || *lastLog.Message != "This is an error message" {
		t.Errorf("Expected error message 'This is an error message', got '%v'", lastLog.Message)
	}

	if lastLog.RequestPath == nil || *lastLog.RequestPath != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%v'", lastLog.RequestPath)
	}

	if lastLog.RequestMethod == nil || *lastLog.RequestMethod != "GET" {
		t.Errorf("Expected method 'GET', got '%v'", lastLog.RequestMethod)
	}

	if lastLog.HttpStatus == nil || *lastLog.HttpStatus != 500 {
		t.Errorf("Expected status code 500, got %v", lastLog.HttpStatus)
	}

	if lastLog.Url == nil || *lastLog.Url != "https://example.com/test" {
		t.Errorf("Expected url 'https://example.com/test', got '%v'", lastLog.Url)
	}

	if lastLog.Hostname == nil || *lastLog.Hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%v'", lastLog.Hostname)
	}

	t.Logf("Successfully logged error to database with ID: %d", lastLog.ID)
}
