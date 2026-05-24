// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAddExeSuffix(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{"No suffix", "ripgrep", "ripgrep.exe"},
		{"With suffix", "ripgrep.exe", "ripgrep.exe"},
		{"Empty string", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := addExeSuffix(test.input)

			if result != test.expected {
				t.Errorf("addExeSuffix failed, got %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestAddExeSuffixes(t *testing.T) {
	input := []string{"a", "b", "c.exe"}
	expected := []string{"a.exe", "b.exe", "c.exe"}

	result := addExeSuffixes(input)
	if !slices.Equal(result, expected) {
		t.Errorf("addExeSuffixes failed, got %v, expected %v", result, expected)
	}
}

func TestStripExeSuffix(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{"No suffix", "ripgrep", "ripgrep"},
		{"With suffix", "ripgrep.exe", "ripgrep"},
		{"Empty string", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := stripExeSuffix(test.input)

			if result != test.expected {
				t.Errorf("stripExeSuffix failed, got %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestStripExeSuffixes(t *testing.T) {
	input := []string{"a", "b", "c.exe"}
	expected := []string{"a", "b", "c"}

	result := stripExeSuffixes(input)
	if !slices.Equal(result, expected) {
		t.Errorf("stripExeSuffixes failed, got %v, expected %v", result, expected)
	}
}

func TestGetCacheFilePath(t *testing.T) {
	t.Run("Tooli environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CACHE_DIRECTORY", "/tooli/cache")

		path, err := getCacheFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join("/tooli/cache", cacheFileName)
		if path != expected {
			t.Errorf("invalid cache path: got %q but expected %q", path, expected)
		}
	})

	t.Run("XDG environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CACHE_DIRECTORY", "")
		t.Setenv("XDG_CACHE_HOME", "/mycache")

		path, err := getCacheFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join("/mycache", appName, cacheFileName)
		if path != expected {
			t.Errorf("invalid cache path: got %q but expected %q", path, expected)
		}
	})

	t.Run("No environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CACHE_DIRECTORY", "")
		t.Setenv("XDG_CACHE_HOME", "")

		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			t.Fatalf("unexpected error obtaining user cache directory: %v", err)
		}

		path, err := getCacheFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join(userCacheDir, appName, cacheFileName)
		if path != expected {
			t.Errorf("invalid cache path: got %q but expected %q", path, expected)
		}
	})
}

func TestGetConfigFilePath(t *testing.T) {
	t.Run("Tooli environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CONFIG_DIRECTORY", "/tooli/config")

		path, err := getConfigFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join("/tooli/config", configFileName)
		if path != expected {
			t.Errorf("invalid config path: got %q but expected %q", path, expected)
		}
	})

	t.Run("XDG environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CONFIG_DIRECTORY", "")
		t.Setenv("XDG_CONFIG_HOME", "/myconfig")

		path, err := getConfigFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join("/myconfig", appName, configFileName)
		if path != expected {
			t.Errorf("invalid config path: got %q but expected %q", path, expected)
		}
	})

	t.Run("No environment variable set", func(t *testing.T) {
		t.Setenv("TOOLI_CONFIG_DIRECTORY", "")
		t.Setenv("XDG_CONFIG_HOME", "")

		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("unexpected error obtaining user config directory: %v", err)
		}

		path, err := getConfigFilePath()
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		expected := filepath.Join(userConfigDir, appName, configFileName)
		if path != expected {
			t.Errorf("invalid config path: got %q but expected %q", path, expected)
		}
	})
}

func TestMakeOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()

	testOutputDir := filepath.Join(tempDir, "testa", "testb", "testc")

	err := makeOutputDirectory(testOutputDir)
	if err != nil {
		t.Fatalf("unexpected error creating test output directory: %v", err)
	}

	_, err = os.Stat(testOutputDir)
	if err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	err = makeOutputDirectory(testOutputDir)
	if err != nil {
		t.Errorf("unexpected error when attempting to recreate test output directory: %v", err)
	}
}
