package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiskCache provides file-based caching with MD5-based sharding.
type DiskCache struct {
	Root string
}

// NewDiskCache creates a new DiskCache instance with the specified root directory.
func NewDiskCache(root string) *DiskCache {
	return &DiskCache{Root: root}
}

// Find attempts to retrieve a cached file for the given key.
// Returns nil, nil for a cache miss (not an error).
func (c *DiskCache) Find(key string) ([]byte, error) {
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
func (c *DiskCache) Write(key string, data []byte) error {
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
func (c *DiskCache) Delete(key string) error {
	cachePath := c.buildPath(key)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}
	return os.Remove(absPath)
}

// buildPath shards files using the first two characters of the MD5
// to prevent too many files in one directory.
func (c *DiskCache) buildPath(key string) string {
	md5 := MD5(key)
	return filepath.Join(c.Root, md5[:2], md5)
}
