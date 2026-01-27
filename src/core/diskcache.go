package core

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

// DiskCache provides file-based caching with MD5-based sharding.
type DiskCache struct {
	Root    string
	TTL     time.Duration
	MaxSize int64
}

// Option configures the DiskCache.
type Option func(*DiskCache)

// WithTTL sets the time-to-live for cached items.
func WithTTL(ttl time.Duration) Option {
	return func(c *DiskCache) {
		c.TTL = ttl
	}
}

// WithMaxSize sets the maximum size of the cache in bytes.
func WithMaxSize(size int64) Option {
	return func(c *DiskCache) {
		c.MaxSize = size
	}
}

// NewDiskCache creates a new DiskCache instance with the specified root directory.
func NewDiskCache(root string, opts ...Option) *DiskCache {
	c := &DiskCache{
		Root:    root,
		MaxSize: 1 * 1024 * 1024 * 1024, // Default 1GB
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Find attempts to retrieve a cached file for the given key.
// Returns nil, nil for a cache miss (not an error).
func (c *DiskCache) Find(key string) ([]byte, error) {
	cachePath := c.buildPath(key)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	exists := err == nil
	if !exists {
		return nil, nil // A cache miss is not an error.
	}

	// Check TTL
	if c.TTL > 0 {
		if time.Since(info.ModTime()) > c.TTL {
			_ = os.Remove(absPath) // Remove expired item
			return nil, nil        // Treat as cache miss
		}
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

// pruningFile represents a file in the cache for pruning purposes.
type pruningFile struct {
	path    string
	size    int64
	modTime time.Time
}

// Prune enforces the MaxSize limit by removing oldest items.
func (c *DiskCache) Prune() error {
	if c.MaxSize <= 0 {
		return nil
	}

	var files []pruningFile
	var totalSize int64

	err := filepath.WalkDir(c.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If we can't read a directory/file, just skip it but don't fail the whole prune
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			size := info.Size()
			totalSize += size
			files = append(files, pruningFile{
				path:    path,
				size:    size,
				modTime: info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking cache dir: %w", err)
	}

	if totalSize <= c.MaxSize {
		slog.Info("no need to prune",
			"root", filepath.Base(c.Root),
			"size", humanize.Bytes(uint64(totalSize)),
			"limit", humanize.Bytes(uint64(c.MaxSize)),
			"ttl", c.TTL,
		)
		return nil
	}

	// Sort by modification time, oldest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	for _, f := range files {
		if totalSize <= c.MaxSize {
			break
		}
		err := os.Remove(f.path)
		if err == nil {
			totalSize -= f.size
		}
	}

	return nil
}
