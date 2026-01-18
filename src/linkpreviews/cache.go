package linkpreviews

import (
	"fmt"
	"os"
	"path/filepath"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/core"
)

var linkPreviewCacheDir string

func InitCache() {
	linkPreviewCacheDir = filepath.Join(conf.Config.DataDir, "cache", "link-preview")
}

// findCached attempts to retrieve a cached PNG image for the given URL and selector.
// An [err] return means the cache has an issue; it does not mean the lookup failed.
func findCached(url, selector string) (png []byte, err error) {
	cachePath := buildCachePath(url, selector)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return nil, err
	}

	exists, err := core.FileExists(absPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil // A cache miss is not an error.
	}

	cached, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("error reading cache: %w", err)
	}

	return cached, nil
}

// Shard files using the first two characters of the MD5 to prevent too many files in one directory.
func buildCachePath(url, selector string) string {
	md5 := core.MD5(url + selector)
	return filepath.Join(linkPreviewCacheDir, md5[:2], md5)
}

func writeToCache(url, selector string, png []byte) (err error) {
	cachePath := buildCachePath(url, selector)
	absPath, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}

	f, err := core.CreateFile(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(png); err != nil {
		return err
	}
	f.Sync()
	return nil
}
