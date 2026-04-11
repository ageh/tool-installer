// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllBinariesExist(t *testing.T) {
	tempDir := t.TempDir()

	app := App{}
	tool := Tool{
		Binaries: []Binary{
			{Name: "rg"},
			{Name: "fd", RenameTo: "finder"},
		},
	}

	t.Run("All binaries present", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tempDir, "rg"), []byte("binary"), 0755)
		if err != nil {
			t.Fatalf("unexpected error creating test binary: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, "finder"), []byte("binary"), 0755)
		if err != nil {
			t.Fatalf("unexpected error creating renamed test binary: %v", err)
		}

		found, err := app.allBinariesExist(tempDir, tool)
		if err != nil {
			t.Fatalf("unexpected error checking binaries: %v", err)
		}

		if !found {
			t.Fatal("expected all binaries to be found")
		}
	})

	t.Run("Missing binary returns false", func(t *testing.T) {
		err := os.Remove(filepath.Join(tempDir, "finder"))
		if err != nil {
			t.Fatalf("unexpected error removing test binary: %v", err)
		}

		found, err := app.allBinariesExist(tempDir, tool)
		if err != nil {
			t.Fatalf("unexpected error checking binaries: %v", err)
		}

		if found {
			t.Fatal("expected missing binary to be detected")
		}
	})
}

func TestToolsFromCacheSkipsStaleEntries(t *testing.T) {
	tempDir := t.TempDir()

	app := App{
		config: Configuration{
			InstallationDirectory: tempDir,
			Tools: map[string]Tool{
				"ripgrep": {Binaries: []Binary{{Name: "rg"}}},
				"fd":      {Binaries: []Binary{{Name: "fd"}}},
			},
		},
		cache: Cache{
			Tools: map[string]string{
				"ripgrep": "1.0.0",
				"fd":      "2.0.0",
				"ghost":   "3.0.0",
			},
		},
	}

	err := os.WriteFile(filepath.Join(tempDir, "rg"), []byte("binary"), 0755)
	if err != nil {
		t.Fatalf("unexpected error creating installed binary: %v", err)
	}

	tools, notFound, stale, err := app.toolsFromCache()
	if err != nil {
		t.Fatalf("unexpected error reading tools from cache: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("expected one installed tool, got %d", len(tools))
	}

	if _, found := tools["ripgrep"]; !found {
		t.Fatal("expected ripgrep to be treated as installed")
	}

	if len(notFound) != 1 || notFound[0] != "ghost" {
		t.Fatalf("unexpected cache-only tools: %v", notFound)
	}

	if len(stale) != 1 || stale[0] != "fd" {
		t.Fatalf("unexpected stale cache entries: %v", stale)
	}
}
