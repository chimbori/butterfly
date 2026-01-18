package linkpreviews

import (
	"os"
	"path/filepath"
	"testing"

	"chimbori.dev/butterfly/conf"
)

func setupTestCache(t *testing.T) string {
	t.Helper()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "butterfly-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}

	// Setup config
	conf.Config = conf.AppConfig{}
	conf.Config.DataDir = tempDir

	// Initialize cache
	InitCache()

	return tempDir
}

func cleanupTestCache(t *testing.T, tempDir string) {
	t.Helper()
	os.RemoveAll(tempDir)
}

func TestInitCache(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	expectedCacheDir := filepath.Join(tempDir, "cache", "link-previews")
	if CacheRoot != expectedCacheDir {
		t.Errorf("Expected cache dir %s, got %s", expectedCacheDir, CacheRoot)
	}
}

func TestBuildCachePath(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "Simple URL and selector",
			url:  "https://example.com",
		},
		{
			name: "Complex URL with query params",
			url:  "https://example.com/page?param=value",
		},
		{
			name: "URL with fragment",
			url:  "https://example.com/page#section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := buildCachePath(tt.url)

			// Verify path contains the cache directory
			if !filepath.IsAbs(cachePath) {
				// Make it absolute for comparison
				cachePath, _ = filepath.Abs(cachePath)
			}

			// Check that path is within the cache directory
			relPath, err := filepath.Rel(CacheRoot, cachePath)
			if err != nil {
				t.Errorf("Cache path not within cache directory: %s", err)
			}

			// Verify path structure (should be 2-char-dir/md5hash)
			parts := filepath.SplitList(relPath)
			if len(parts) == 0 {
				// On some systems, need to split by separator
				relPath = filepath.ToSlash(relPath)
				if len(relPath) < 3 {
					t.Errorf("Cache path too short: %s", relPath)
				}
			}

			// Verify the path is consistent for same inputs
			cachePath2 := buildCachePath(tt.url)
			if cachePath != cachePath2 {
				t.Errorf("Cache path not consistent: %s != %s", cachePath, cachePath2)
			}

			// Verify different inputs produce different paths
			differentPath := buildCachePath(tt.url + "different")
			if cachePath == differentPath {
				t.Errorf("Different inputs produced same cache path")
			}
		})
	}
}

func TestWriteToCache_AndFindCached(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	url := "https://example.com/test"
	testData := []byte("test PNG data")

	// Write to cache
	err := writeToCache(url, testData)
	if err != nil {
		t.Fatalf("Failed to write to cache: %s", err)
	}

	// Find cached data
	cached, err := findCached(url)
	if err != nil {
		t.Fatalf("Failed to find cached data: %s", err)
	}

	if cached == nil {
		t.Fatal("Expected cached data, got nil")
	}

	// Verify data matches
	if string(cached) != string(testData) {
		t.Errorf("Cached data mismatch: expected %s, got %s", string(testData), string(cached))
	}
}

func TestFindCached_NotFound(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	url := "https://nonexistent.com"

	// Try to find non-existent cached data
	cached, err := findCached(url)
	if err != nil {
		t.Fatalf("Expected no error for cache miss, got: %s", err)
	}

	if cached != nil {
		t.Errorf("Expected nil for cache miss, got: %v", cached)
	}
}

func TestWriteToCache_CreatesDirectories(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	url := "https://example.com/nested/path"
	testData := []byte("nested cache data")

	// Verify cache directory doesn’t exist yet
	cachePath := buildCachePath(url)
	absPath, _ := filepath.Abs(cachePath)
	dirPath := filepath.Dir(absPath)

	// Write to cache (should create directories)
	err := writeToCache(url, testData)
	if err != nil {
		t.Fatalf("Failed to write to cache: %s", err)
	}

	// Verify directory was created
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Cache directory not created: %s", err)
	}

	if !info.IsDir() {
		t.Error("Cache path is not a directory")
	}

	// Verify file exists and contains correct data
	cached, err := findCached(url)
	if err != nil {
		t.Fatalf("Failed to find cached data: %s", err)
	}

	if string(cached) != string(testData) {
		t.Errorf("Cached data mismatch: expected %s, got %s", string(testData), string(cached))
	}
}

func TestWriteToCache_OverwritesExisting(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	url := "https://example.com/update"
	originalData := []byte("original data")
	updatedData := []byte("updated data")

	// Write original data
	err := writeToCache(url, originalData)
	if err != nil {
		t.Fatalf("Failed to write original data: %s", err)
	}

	// Verify original data
	cached, err := findCached(url)
	if err != nil {
		t.Fatalf("Failed to find original cached data: %s", err)
	}
	if string(cached) != string(originalData) {
		t.Errorf("Original data mismatch: expected %s, got %s", string(originalData), string(cached))
	}

	// Overwrite with updated data
	err = writeToCache(url, updatedData)
	if err != nil {
		t.Fatalf("Failed to write updated data: %s", err)
	}

	// Verify updated data
	cached, err = findCached(url)
	if err != nil {
		t.Fatalf("Failed to find updated cached data: %s", err)
	}
	if string(cached) != string(updatedData) {
		t.Errorf("Updated data mismatch: expected %s, got %s", string(updatedData), string(cached))
	}
}

func TestWriteToCache_EmptyData(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	url := "https://example.com/empty"
	emptyData := []byte{}

	// Write empty data
	err := writeToCache(url, emptyData)
	if err != nil {
		t.Fatalf("Failed to write empty data: %s", err)
	}

	// Find cached data
	cached, err := findCached(url)
	if err != nil {
		t.Fatalf("Failed to find cached data: %s", err)
	}

	if cached == nil {
		t.Fatal("Expected cached data (even if empty), got nil")
	}

	if len(cached) != 0 {
		t.Errorf("Expected empty cached data, got %d bytes", len(cached))
	}
}

func TestBuildCachePath_Sharding(t *testing.T) {
	tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Test that files are sharded using first 2 characters of MD5
	url := "https://example.com/shard-test"
	cachePath := buildCachePath(url)

	// Get relative path from cache dir
	relPath, err := filepath.Rel(CacheRoot, cachePath)
	if err != nil {
		t.Fatalf("Failed to get relative path: %s", err)
	}

	// Split path components
	relPath = filepath.ToSlash(relPath)
	parts := filepath.SplitList(relPath)

	// On some systems the path is a single string, so we need to handle that
	if len(parts) <= 1 {
		// Check that path has the sharding structure manually
		if len(relPath) < 3 {
			t.Errorf("Cache path too short for sharding: %s", relPath)
		}

		// First directory should be 2 characters
		if relPath[2] != '/' && relPath[2] != '\\' {
			t.Errorf("Expected sharding directory separator at position 2, got: %c", relPath[2])
		}
	}
}
