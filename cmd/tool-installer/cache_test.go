// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheWriteRead(t *testing.T) {
	tempDir := t.TempDir()

	t.Setenv("TOOLI_CACHE_DIRECTORY", tempDir)

	t.Run("Missing file produces empty cache", func(t *testing.T) {
		cache, err := getCache()
		if err != nil {
			t.Fatalf("unexpected error trying to read cache from non-existent file: %v", err)
		}

		if len(cache.Tools) != 0 {
			t.Errorf("unexpected amount of tools in new cache: got %d, expected 0", len(cache.Tools))
		}
	})

	t.Run("Added entry survives writing and reading", func(t *testing.T) {
		cache, err := getCache()
		if err != nil {
			t.Fatalf("unexpected error trying to read cache: %v", err)
		}

		cache.add("tooli", "1.0.0")

		err = cache.writeCache()
		if err != nil {
			t.Fatalf("unexpected error writing cache: %v", err)
		}

		cache, err = getCache()
		if err != nil {
			t.Fatalf("unexpected error trying to read cache: %v", err)
		}

		if len(cache.Tools) != 1 {
			t.Fatalf("expected to find one entry in the cache but got %d", len(cache.Tools))
		}

		if !cache.contains("tooli") {
			t.Fatal("expected to find 'tooli' in the cache")
		}

		if cache.Tools["tooli"] != "1.0.0" {
			t.Errorf("expected to find version '1.0.0' in cache but got %q", cache.Tools["tooli"])
		}
	})

	t.Run("Explicit null tools map does not panic", func(t *testing.T) {
		cachePath := filepath.Join(tempDir, cacheFileName)
		if err := os.WriteFile(cachePath, []byte(`{"tools": null}`), 0o644); err != nil {
			t.Fatalf("unexpected error writing test cache file: %v", err)
		}

		cache, err := getCache()
		if err != nil {
			t.Fatalf("unexpected error trying to read cache with null tools: %v", err)
		}

		cache.add("tooli", "1.0.0")

		if !cache.contains("tooli") {
			t.Error("expected to find 'tooli' in the cache after adding it")
		}
	})
}
