package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)

	if cache.Root != root {
		t.Errorf("Expected root %s, got %s", root, cache.Root)
	}
}

func TestCacheWrite(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "test-key"
	data := []byte("test data")

	err := cache.Write(key, data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	cachePath := cache.buildPath(key)
	absPath, _ := filepath.Abs(cachePath)
	if _, err := os.Stat(absPath); err != nil {
		t.Errorf("Cache file not found: %v", err)
	}
}

func TestCacheFind_Hit(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "test-key"
	data := []byte("test data")

	cache.Write(key, data)

	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if string(found) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(found))
	}
}

func TestCacheFind_Miss(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "nonexistent-key"

	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find should not error on miss: %v", err)
	}

	if found != nil {
		t.Errorf("Expected nil for cache miss, got %v", found)
	}
}

func TestCacheDelete(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "test-key"
	data := []byte("test data")

	cache.Write(key, data)

	err := cache.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify file is deleted
	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find after delete failed: %v", err)
	}

	if found != nil {
		t.Errorf("Expected nil after delete, got %v", found)
	}
}

func TestCacheBuildPath_Sharding(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "test-key"

	path := cache.buildPath(key)
	md5 := MD5(key)

	// Verify sharding structure
	expected := filepath.Join(root, md5[:2], md5)
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestCacheMultipleKeys(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)

	keys := []string{"key1", "key2", "key3"}
	dataMap := make(map[string][]byte)

	// Write multiple keys
	for i, key := range keys {
		data := []byte("data" + string(rune(i)))
		dataMap[key] = data
		if err := cache.Write(key, data); err != nil {
			t.Fatalf("Write failed for %s: %v", key, err)
		}
	}

	// Verify all keys can be found
	for key, expectedData := range dataMap {
		found, err := cache.Find(key)
		if err != nil {
			t.Fatalf("Find failed for %s: %v", key, err)
		}
		if string(found) != string(expectedData) {
			t.Errorf("Mismatch for %s: expected %s, got %s", key, string(expectedData), string(found))
		}
	}
}

func TestCacheWriteEmptyData(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "empty-key"
	data := []byte{}

	err := cache.Write(key, data)
	if err != nil {
		t.Fatalf("Write empty data failed: %v", err)
	}

	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find empty data failed: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("Expected empty data, got %v", found)
	}
}

func TestCacheWriteLargeData(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "large-key"
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	err := cache.Write(key, data)
	if err != nil {
		t.Fatalf("Write large data failed: %v", err)
	}

	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find large data failed: %v", err)
	}

	if len(found) != len(data) {
		t.Errorf("Expected %d bytes, got %d", len(data), len(found))
	}
}

func TestCacheDeleteNonexistent(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "nonexistent-key"

	err := cache.Delete(key)
	if err == nil {
		t.Errorf("Expected error when deleting nonexistent key")
	}
}

func TestCacheOverwrite(t *testing.T) {
	root := t.TempDir()
	cache := NewCache(root)
	key := "test-key"
	data1 := []byte("original data")
	data2 := []byte("new data")

	cache.Write(key, data1)
	cache.Write(key, data2)

	found, err := cache.Find(key)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if string(found) != string(data2) {
		t.Errorf("Expected %s, got %s", string(data2), string(found))
	}
}
