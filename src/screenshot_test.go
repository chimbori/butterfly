package main

import (
	"context"
	"testing"
	"time"
)

func TestTakeScreenshot_ValidPageVisibleElement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := "data:text/html,<html><body><div id='content' style='width:100px;height:100px;background:red;'>Test Content</div></body></html>"
	selector := "#content"

	screenshot, err := takeScreenshot(ctx, url, selector)
	if err != nil {
		t.Fatalf("Expected no error for valid page and selector, got: %s", err.Error())
	}

	if len(screenshot) == 0 {
		t.Fatal("Expected non-empty screenshot data")
	}

	assertValidPNG(t, screenshot)
}

func TestTakeScreenshot_ValidPageWithMultipleElements(t *testing.T) {
	// This test requires chromedp to be available
	// Skip if running in environments without Chrome/Chromium
	if testing.Short() {
		t.Skip("Skipping test that requires Chrome/Chromium in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// HTML with multiple elements - using inline styles for simplicity
	url := "data:text/html,<html><body><div id='header' style='width:200px;height:50px;background:blue;'>Header</div><div id='content' style='width:200px;height:100px;background:green;'>Main Content</div><div id='footer' style='width:200px;height:50px;background:red;'>Footer</div></body></html>"
	selector := "#content"

	screenshot, err := takeScreenshot(ctx, url, selector)
	if err != nil {
		t.Fatalf("Expected no error for valid page with multiple elements, got: %s", err.Error())
	}

	if len(screenshot) == 0 {
		t.Fatal("Expected non-empty screenshot data")
	}

	assertValidPNG(t, screenshot)
}

func TestTakeScreenshot_HiddenElement(t *testing.T) {
	// This test verifies that hidden elements are made visible before screenshot
	// Skip if running in environments without Chrome/Chromium
	if testing.Short() {
		t.Skip("Skipping test that requires Chrome/Chromium in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// HTML with a hidden element that should be made visible by the screenshot function
	// Using inline style display:none to test the unhiding functionality
	url := "data:text/html,<html><body><div id='content' style='width:200px;height:100px;background:purple;display:none;'>Hidden Content</div></body></html>"
	selector := "#content"

	screenshot, err := takeScreenshot(ctx, url, selector)
	if err != nil {
		t.Fatalf("Expected no error for hidden element (should be made visible), got: %s", err.Error())
	}

	if len(screenshot) == 0 {
		t.Fatal("Expected non-empty screenshot data for hidden element made visible")
	}

	assertValidPNG(t, screenshot)
}

func TestTakeScreenshot_MissingSelector(t *testing.T) {
	ctx := context.Background()
	url := "https://example.com"
	selector := ""

	_, err := takeScreenshot(ctx, url, selector)
	if err == nil {
		t.Error("Expected error for missing selector")
	}
	if err.Error() != "missing selector" {
		t.Errorf("Expected 'missing selector' error, got: %s", err.Error())
	}
}

func TestTakeScreenshot_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := "not-a-valid-url"
	selector := "body"

	_, err := takeScreenshot(ctx, url, selector)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestTakeScreenshot_NonExistentSelector(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a simple HTML data URL
	url := "data:text/html,<html><body><div id='content'>Hello</div></body></html>"
	selector := "#non-existent-element"

	_, err := takeScreenshot(ctx, url, selector)
	if err == nil {
		t.Error("Expected error for non-existent selector")
	}
}

func TestTakeScreenshot_ContextCancellation(t *testing.T) {
	// Test that context cancellation is handled properly
	if testing.Short() {
		t.Skip("Skipping test that requires Chrome/Chromium in short mode")
	}

	// Create a context that’s already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	url := "data:text/html,<html><body><div id='content'>Test</div></body></html>"
	selector := "#content"

	_, err := takeScreenshot(ctx, url, selector)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

// assertValidPNG checks that the given byte slice is a valid PNG file
// by verifying it starts with the PNG magic bytes: 137 80 78 71 13 10 26 10
func assertValidPNG(t *testing.T, data []byte) {
	t.Helper()

	if len(data) < 8 {
		t.Fatalf("PNG data too small (%d bytes), expected at least 8 bytes", len(data))
	}

	// PNG magic bytes: 137 80 78 71 13 10 26 10
	if data[0] != 137 || data[1] != 80 || data[2] != 78 || data[3] != 71 {
		t.Errorf("Invalid PNG magic bytes: got [%d %d %d %d], expected [137 80 78 71]",
			data[0], data[1], data[2], data[3])
	}
}
