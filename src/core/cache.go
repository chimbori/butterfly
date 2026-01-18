package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cache provides file-based caching with MD5-based sharding.
type Cache struct {
	Root string
}

// NewCache creates a new Cache instance with the specified root directory.
func NewCache(root string) *Cache {
	return &Cache{Root: root}
}

// Find attempts to retrieve a cached file for the given key.
// Returns nil, nil for a cache miss (not an error).
func (c *Cache) Find(key string) ([]byte, error) {
	cachePath := c.buildPath(key)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	exists := err == nil
	if !exists {
		return nil, nil // A cache miss is not an error.
	}

	cached, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("error reading cache: %w", err)
	}

	return cached, nil
}

// Write stores data in the cache for the given key (URL).
func (c *Cache) Write(key string, data []byte) error {
	cachePath := c.buildPath(key)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}

	f, err := CreateFile(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return err
	}
	f.Sync()
	return nil
}

// Delete removes a cached file for the given key (URL).
func (c *Cache) Delete(key string) error {
	cachePath := c.buildPath(key)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}
	return os.Remove(absPath)
}

// buildPath shards files using the first two characters of the MD5
// to prevent too many files in one directory.
func (c *Cache) buildPath(key string) string {
	md5 := MD5(key)
	return filepath.Join(c.Root, md5[:2], md5)
}
