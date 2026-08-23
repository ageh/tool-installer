// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestAllBinariesExist(t *testing.T) {
	tempDir := t.TempDir()

	app := App{}
	tool := Tool{
		Binaries: []Binary{
			{Name: "rg"},
			{Name: "fd"},
		},
	}

	t.Run("All binaries present", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tempDir, "rg"), []byte("binary"), 0o755)
		if err != nil {
			t.Fatalf("unexpected error creating test binary rg: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, "fd"), []byte("binary"), 0o755)
		if err != nil {
			t.Fatalf("unexpected error creating test binary fd: %v", err)
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
		err := os.Remove(filepath.Join(tempDir, "fd"))
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

	t.Run("Directory at binary path is not treated as installed", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tempDir, "fd"), []byte("binary"), 0o755)
		if err != nil {
			t.Fatalf("unexpected error creating test binary fd: %v", err)
		}

		err = os.Remove(filepath.Join(tempDir, "rg"))
		if err != nil {
			t.Fatalf("unexpected error removing test binary: %v", err)
		}

		err = os.Mkdir(filepath.Join(tempDir, "rg"), 0o755)
		if err != nil {
			t.Fatalf("unexpected error creating directory at binary path: %v", err)
		}

		found, err := app.allBinariesExist(tempDir, tool)
		if err != nil {
			t.Fatalf("unexpected error checking binaries: %v", err)
		}

		if found {
			t.Fatal("expected a directory at the binary path to not count as installed")
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

	err := os.WriteFile(filepath.Join(tempDir, "rg"), []byte("binary"), 0o755)
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

func TestSaveAddedTool(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	originalTool := validTestTool("rg")
	app := App{
		config: Configuration{
			Version:               currentConfigurationVersion,
			InstallationDirectory: "~/.local/bin",
			Tools: map[string]Tool{
				"ripgrep": originalTool,
			},
		},
		configLocation: configPath,
	}
	originalConfig := app.config
	originalConfig.Tools = maps.Clone(app.config.Tools)

	err := app.saveAddedTool("duplicate", validTestTool("RG"))
	if err == nil {
		t.Fatal("expected adding a conflicting tool to fail")
	}

	if !reflect.DeepEqual(app.config, originalConfig) {
		t.Errorf("configuration changed after rejecting conflicting tool: got %+v, expected %+v", app.config, originalConfig)
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Errorf("configuration file was written after rejecting conflicting tool: %v", err)
	}
}

func TestInstallToolsReinstallsMissingBinaries(t *testing.T) {
	var assetHit atomic.Bool

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/tool/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.0.0",
			Assets:  []Asset{{Name: "tool.bin", Url: gitHubApiUrl + "/assets/tool"}},
		})
	})
	mux.HandleFunc("/assets/tool", func(w http.ResponseWriter, r *http.Request) {
		assetHit.Store(true)
		_, _ = w.Write([]byte("binary-data"))
	})

	tempDir := t.TempDir()

	app := App{
		downloader: newTestDownloader(t, mux),
		config: Configuration{
			InstallationDirectory: tempDir,
			Tools: map[string]Tool{
				"tool": testDownloadableTool("owner", "tool"),
			},
		},
		cache: Cache{Tools: map[string]string{"tool": "v1.0.0"}},
	}

	if err := app.installTools(nil); err != nil {
		t.Fatalf("installTools failed: %v", err)
	}

	if !assetHit.Load() {
		t.Fatal("expected the asset to be re-downloaded even though the cache reported the tool as up to date")
	}

	if _, err := os.Stat(filepath.Join(tempDir, "tool")); err != nil {
		t.Fatalf("expected the binary to be installed on disk: %v", err)
	}
}
