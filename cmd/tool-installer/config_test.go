// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"testing"
)

const testLegacyConfig = `{
	"install_dir": "~/.local/bin",
	"tools": {
		"ripgrep": {
			"owner": "BurntSushi",
			"repository": "ripgrep",
			"linux_asset": "linux\\.tar\\.gz$",
			"windows_asset": "windows\\.zip$",
			"description": "Fast grep",
			"binaries": [
				{
					"name":"rg"
				}
			]
		}
	}
}`

const testFutureConfig = `{
	"version": 999999999,
	"install_dir": "~/.local/bin",
	"tools": {
		"ripgrep": {
			"owner": "BurntSushi",
			"repository": "ripgrep",
			"linux_asset": "linux\\.tar\\.gz$",
			"description": "Fast grep",
			"binaries": [
				{
					"name":"rg"
				}
			]
		}
	}
}`

const testConfig = `{
	"version": 3,
	"install_dir": "~/.local/bin",
	"github_token": "configured-token",
	"tools": {
		"ripgrep": {
			"owner": "BurntSushi",
			"repository": "ripgrep",
			"asset": "linux\\.tar\\.gz$",
			"description": "Fast grep",
			"binaries": [
				{
					"name":"rg"
				}
			]
		}
	}
}`

const testInvalidConfig = `{
	"version": 3,
	"install_dir": "~/.local/bin",
	"github_token": "configured-token",
	"tools": {
		"ripgrep": {
			"owner": "BurntSushi",
			"repository": "ripgrep",
			"asset": "linux\\.tar\\.gz$",
			"description": "Fast grep",
			"binaries": [
				{
					"name":"rg"
				}
			]
		},
		"fd": {
			"owner": "sharkdp",
			"repository": "fd",
			"asset": "linux\\.tar\\.gz$",
			"description": "Fast find",
			"binaries": [
				{
					"name":"RG"
				}
			]
		}
	}
}`

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("unexpected error creating file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		t.Fatalf("unexpected error writing to file: %v", err)
	}
}

func validTestTool(binaryName string) Tool {
	pattern := `linux\.tar\.gz$`
	return Tool{
		Owner:      "owner",
		Repository: "repository",
		Asset: AssetRegex{
			Pattern: pattern,
			Regex:   regexp.MustCompile(pattern),
		},
		Binaries: []Binary{{Name: binaryName}},
	}
}

func validTestConfig() Configuration {
	return Configuration{
		InstallationDirectory: "~/.local/bin",
		Tools: map[string]Tool{
			"tool": validTestTool("tool"),
		},
	}
}

func TestBinaryMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    Binary
		expected string
	}{
		{name: "No source_names", input: Binary{Name: "rg"}, expected: `{"name":"rg"}`},
		{name: "Empty source_names", input: Binary{Name: "fd", SourceNames: []string{}}, expected: `{"name":"fd"}`},
		{name: "Non-empty source_names", input: Binary{Name: "fd", SourceNames: []string{"find"}}, expected: `{"name":"fd","source_names":["find"]}`},
		{name: "Windows exe suffix stripped", input: Binary{Name: "rg.exe"}, expected: `{"name":"rg"}`},
		{name: "Non-empty platforms", input: Binary{Name: "bwrap", Platforms: []string{"linux"}}, expected: `{"name":"bwrap","platforms":["linux"]}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := json.Marshal(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != test.expected {
				t.Errorf("got %q, expected %q", got, test.expected)
			}
		})
	}
}

func TestBinaryUnmarshalJSON(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`not json`), &b); err == nil {
			t.Error("expected an error for invalid JSON, got nil")
		}
	})

	tests := []struct {
		name     string
		input    string
		expected Binary
	}{
		{"No rename_to", `{"name": "rg"}`, Binary{Name: "rg"}},
		{"Empty rename_to", `{"name": "rg", "rename_to": ""}`, Binary{Name: "rg"}},
		{"Non-empty rename_to", `{"name": "rg", "rename_to": "ripgrep"}`, Binary{Name: "rg", RenameTo: "ripgrep"}},
		{"OS-only platforms entry", `{"name": "bwrap", "platforms": ["linux"]}`, Binary{Name: "bwrap", Platforms: []string{"linux"}}},
		{"OS/arch platforms entry", `{"name": "sandbox-setup", "platforms": ["windows/amd64"]}`, Binary{Name: "sandbox-setup", Platforms: []string{"windows/amd64"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var b Binary
			if err := json.Unmarshal([]byte(test.input), &b); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if b.Name != test.expected.Name {
				t.Errorf("got name %q, expected %q", b.Name, test.expected.Name)
			}

			if b.RenameTo != test.expected.RenameTo {
				t.Errorf("got rename_to %q, expected %q", b.RenameTo, test.expected.RenameTo)
			}

			if !slices.Equal(b.Platforms, test.expected.Platforms) {
				t.Errorf("got platforms %v, expected %v", b.Platforms, test.expected.Platforms)
			}
		})
	}

	t.Run("Invalid platforms entry", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "platforms": ["not-a-platform"]}`), &b); err == nil {
			t.Error("expected an error for an invalid platforms entry, got nil")
		}
	})

	t.Run("Invalid name: empty string", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": ""}`), &b); err == nil {
			t.Error("expected an error for empty name, got nil")
		}
	})

	t.Run("Invalid rename_to: contains period", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "."}`), &b); err == nil {
			t.Error("expected an error for rename_to '.', got nil")
		}
	})

	t.Run("Invalid rename_to: path separator", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "rip/grep"}`), &b); err == nil {
			t.Error("expected an error for rename_to with slash, got nil")
		}
	})

	t.Run("Invalid rename_to: Windows path separator", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "rip\\grep"}`), &b); err == nil {
			t.Error("expected an error for rename_to with backslash, got nil")
		}
	})
}

func TestBinaryMarshalUnmarshalJSON(t *testing.T) {
	tests := []Binary{
		{Name: "rg"},
		{Name: "tooli", SourceNames: []string{"tool-installer"}},
		{Name: "bwrap", Platforms: []string{"linux"}},
	}

	for _, original := range tests {
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal failed for %+v: %v", original, err)
		}

		var restored Binary
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("unmarshal failed for %q: %v", data, err)
		}

		if restored.Name != original.Name {
			t.Errorf("Name mismatch: got %q, expected %q", restored.Name, original.Name)
		}

		if !slices.Equal(restored.SourceNames, original.SourceNames) {
			t.Errorf("SourceNames mismatch: got %v, expected %v", restored.SourceNames, original.SourceNames)
		}

		if !slices.Equal(restored.Platforms, original.Platforms) {
			t.Errorf("Platforms mismatch: got %v, expected %v", restored.Platforms, original.Platforms)
		}
	}
}

func TestAssetRegexMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    AssetRegex
		expected string
	}{
		{"Simple pattern", AssetRegex{Pattern: "windows-x86_64\\.tar\\.gz$"}, `"windows-x86_64\\.tar\\.gz$"`},
		{
			"Complex pattern",
			AssetRegex{Pattern: "hugo-extended_v\\d+\\.\\d+\\.\\d+_windows-x86_64\\.zip$"},
			`"hugo-extended_v\\d+\\.\\d+\\.\\d+_windows-x86_64\\.zip$"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := json.Marshal(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != test.expected {
				t.Errorf("got %q, expected %q", got, test.expected)
			}
		})
	}
}

func TestAssetRegexUnmarshalJSON(t *testing.T) {
	t.Run("Invalid JSON: not a string", func(t *testing.T) {
		var a AssetRegex
		if err := json.Unmarshal([]byte(`123`), &a); err == nil {
			t.Error("expected an error for non-string JSON, got nil")
		}
	})

	t.Run("Invalid regex", func(t *testing.T) {
		var a AssetRegex
		if err := json.Unmarshal([]byte(`"[invalid"`), &a); err == nil {
			t.Error("expected an error for invalid regex, got nil")
		}
	})

	t.Run("Empty pattern", func(t *testing.T) {
		var a AssetRegex
		if err := json.Unmarshal([]byte(`""`), &a); err == nil {
			t.Error("expected an error for empty regex, got nil")
		}
	})

	tests := []struct {
		name          string
		input         string
		expected      string
		assetNamePass string
		assetNameFail string
	}{
		{"Simple pattern", `"windows-x86_64\\.zip$"`, "windows-x86_64\\.zip$", "ripgrep-15.0.0-windows-x86_64.zip", "ripgrep-win32.zip"},
		{
			"Complex pattern",
			`"hugo-extended_v\\d+\\.\\d+\\.\\d+_windows-x86_64\\.zip$"`,
			"hugo-extended_v\\d+\\.\\d+\\.\\d+_windows-x86_64\\.zip$",
			"hugo-extended_v1.1.1_windows-x86_64.zip",
			"error.zip",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var a AssetRegex
			if err := json.Unmarshal([]byte(test.input), &a); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if a.Pattern != test.expected {
				t.Errorf("got pattern %q, expected %q", a.Pattern, test.expected)
			}

			if a.Regex == nil {
				t.Error("expected Regex to be compiled, got nil")
			}

			if !a.Regex.MatchString(test.assetNamePass) {
				t.Errorf("unmarshaled regex failed to match %q", test.assetNamePass)
			}

			if a.Regex.MatchString(test.assetNameFail) {
				t.Errorf("unmarshaled regex matched %q even though it should not", test.assetNameFail)
			}
		})
	}
}

func TestAssetRegexMarshalUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{"Simple pattern", `windows-x86_64\.tar\.gz`},
		{"Complex pattern", `fd-v[0-9]+\.[0-9]+\.[0-9]+-x86_64-unknown-linux-musl\.tar\.gz`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := AssetRegex{Pattern: test.pattern}
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var restored AssetRegex
			if err := json.Unmarshal(data, &restored); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if restored.Pattern != original.Pattern {
				t.Errorf("pattern mismatch: got %q, expected %q", restored.Pattern, original.Pattern)
			}
			if restored.Regex == nil {
				t.Error("expected regex to be compiled after unmarshal, got nil")
			}
		})
	}
}

func TestToolForCurrentPlatform(t *testing.T) {
	testTool := Tool{
		Binaries: []Binary{
			{Name: "rg"},
			{Name: "fd"},
			{Name: "bwrap", Platforms: []string{"linux"}},
			{Name: "sandbox-setup", Platforms: []string{"windows/amd64"}},
		},
	}

	tests := []struct {
		name             string
		goos             string
		goarch           string
		expectedBinaries []Binary
	}{
		{"Windows has exe suffix and drops non-matching platforms", "windows", "amd64", []Binary{{Name: "rg.exe"}, {Name: "fd.exe"}, {Name: "sandbox-setup.exe", Platforms: []string{"windows/amd64"}}}},
		{"Windows arm64 drops the amd64-only binary", "windows", "arm64", []Binary{{Name: "rg.exe"}, {Name: "fd.exe"}}},
		{"Non-Windows has no suffix and drops non-matching platforms", "linux", "amd64", []Binary{{Name: "rg"}, {Name: "fd"}, {Name: "bwrap", Platforms: []string{"linux"}}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := testTool.forCurrentPlatform(test.goos, test.goarch)

			if len(result.Binaries) != len(test.expectedBinaries) {
				t.Fatalf("wrong number of binaries: got %d, expected %d: %v", len(result.Binaries), len(test.expectedBinaries), result.Binaries)
			}

			for i := range result.Binaries {
				if result.Binaries[i].Name != test.expectedBinaries[i].Name {
					t.Errorf("wrong Name: got %q, expected %q", result.Binaries[i].Name, test.expectedBinaries[i].Name)
				}

				if result.Binaries[i].RenameTo != test.expectedBinaries[i].RenameTo {
					t.Errorf("wrong RenameTo: got %q, expected %q", result.Binaries[i].RenameTo, test.expectedBinaries[i].RenameTo)
				}

				if !slices.Equal(result.Binaries[i].Platforms, test.expectedBinaries[i].Platforms) {
					t.Errorf("wrong Platforms: got %v, expected %v", result.Binaries[i].Platforms, test.expectedBinaries[i].Platforms)
				}
			}
		})
	}
}

func TestBinaryHasSourceNameForOS(t *testing.T) {
	binary := Binary{Name: "tool.exe", SourceNames: []string{"other.exe"}}

	tests := []struct {
		name     string
		goos     string
		input    string
		expected bool
	}{
		{"Windows matches exact case", "windows", "tool.exe", true},
		{"Windows matches uppercase name", "windows", "Tool.EXE", true},
		{"Windows matches uppercase source name", "windows", "OTHER.EXE", true},
		{"Windows rejects unrelated name", "windows", "unrelated.exe", false},
		{"Non-Windows matches exact case", "linux", "tool.exe", true},
		{"Non-Windows rejects differing case", "linux", "Tool.EXE", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := binary.hasSourceNameForOS(test.input, test.goos)
			if result != test.expected {
				t.Errorf("hasSourceNameForOS(%q, %q) = %v, expected %v", test.input, test.goos, result, test.expected)
			}
		})
	}
}

func TestBinaryAppliesToPlatform(t *testing.T) {
	tests := []struct {
		name      string
		platforms []string
		goos      string
		goarch    string
		expected  bool
	}{
		{"No restriction applies everywhere", nil, "linux", "amd64", true},
		{"OS-only entry matches any arch", []string{"linux"}, "linux", "arm64", true},
		{"OS-only entry rejects other OS", []string{"linux"}, "windows", "amd64", false},
		{"OS/arch entry matches exact arch", []string{"windows/amd64"}, "windows", "amd64", true},
		{"OS/arch entry rejects other arch", []string{"windows/amd64"}, "windows", "arm64", false},
		{"Mixed entries match either", []string{"linux", "windows/amd64"}, "windows", "amd64", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binary := Binary{Name: "tool", Platforms: test.platforms}

			result := binary.appliesToPlatform(test.goos, test.goarch)
			if result != test.expected {
				t.Errorf("appliesToPlatform(%q, %q) = %v, expected %v", test.goos, test.goarch, result, test.expected)
			}
		})
	}
}

func TestGetSanitizedInstallationDirectory(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error obtaining user home directory: %v", err)
	}

	t.Run("Config path tilde", func(t *testing.T) {
		config := Configuration{InstallationDirectory: "~"}

		result, err := config.getSanitizedInstallationDirectory()
		if err != nil {
			t.Fatalf("unexpected error from getSanitizedInstallationDirectory: %v", err)
		}

		expected := homeDir
		if result != expected {
			t.Errorf("got installation directory %q, expected %q", result, expected)
		}
	})

	t.Run("Config path tilde prefix", func(t *testing.T) {
		config := Configuration{InstallationDirectory: "~/.localbin"}

		result, err := config.getSanitizedInstallationDirectory()
		if err != nil {
			t.Fatalf("unexpected error from getSanitizedInstallationDirectory: %v", err)
		}

		expected := filepath.Join(homeDir, ".localbin")
		if result != expected {
			t.Errorf("got installation directory %q, expected %q", result, expected)
		}
	})

	t.Run("Config path other", func(t *testing.T) {
		config := Configuration{InstallationDirectory: "/usr/bin/test"}

		result, err := config.getSanitizedInstallationDirectory()
		if err != nil {
			t.Fatalf("unexpected error from getSanitizedInstallationDirectory: %v", err)
		}

		expected := filepath.Clean("/usr/bin/test")
		if result != expected {
			t.Errorf("got installation directory %q, expected %q", result, expected)
		}
	})
}

func TestMigrateConfiguration(t *testing.T) {
	input := []byte(`{
		"install_dir": "~/.local/bin",
		"tools": {
			"ripgrep": {
				"owner": "BurntSushi",
				"repository": "ripgrep",
				"linux_asset": "linux-x86_64\\.tar\\.gz$",
				"windows_asset": "windows-x86_64\\.zip$",
				"description": "Faster grep",
				"binaries": [
					{
						"name":"rg"
					}
				]
			}
		}
	}`)

	config, err := migrateConfiguration(input, 0)
	if err != nil {
		t.Fatalf("unexpected error in migrateConfiguration: %v", err)
	}

	tool := config.Tools["ripgrep"]

	if config.Version != currentConfigurationVersion {
		t.Fatalf("got version %d, expected %d", config.Version, currentConfigurationVersion)
	}

	switch runtime.GOOS {
	case "windows":
		if tool.Asset.Pattern != "windows-x86_64\\.zip$" {
			t.Fatalf("got unexpected pattern %q", tool.Asset.Pattern)
		}
	case "linux":
		if tool.Asset.Pattern != "linux-x86_64\\.tar\\.gz$" {
			t.Fatalf("got unexpected pattern %q", tool.Asset.Pattern)
		}
	}
}

func TestDefaultConfiguration(t *testing.T) {
	config, err := getDefaultConfiguration()
	if err != nil {
		t.Fatalf("unexpected error getting default configuration: %v", err)
	}

	if config.Version != currentConfigurationVersion {
		t.Errorf("wrong version number: got %d, expected %d", config.Version, currentConfigurationVersion)
	}

	if config.InstallationDirectory != "~/.local/bin" {
		t.Errorf("wrong installation directory: got %q, expected '~/.local/bin'", config.InstallationDirectory)
	}
	if config.GitHubToken != "" {
		t.Errorf("default GitHub token should be empty but got %q", config.GitHubToken)
	}

	for _, name := range defaultTools {
		known, found := knownTools[name]
		if !found {
			t.Errorf("expected known tool entry for default tool %q", name)
			continue
		}

		tool, found := config.Tools[name]
		if !found {
			if _, err := known.intoToolForPlatform(); !errors.Is(err, ErrUnsupportedPlatform) {
				t.Errorf("default tool %q is missing from the default configuration for an unexpected reason: %v", name, err)
			}
			continue
		}

		if tool.Owner == "" {
			t.Errorf("tool %q has an empty string as the owner", name)
		}

		if tool.Repository == "" {
			t.Errorf("tool %q has an empty string as the repository", name)
		}

		if tool.Description == "" {
			t.Errorf("tool %q has an empty string as the description", name)
		}

		if tool.Asset.Pattern == "" {
			t.Errorf("tool %q has an empty string as the asset", name)
		}

		if tool.Asset.Regex == nil {
			t.Errorf("tool %q has no valid compiled asset regex", name)
		}
	}
}

func TestReadConfigurationOrCreateDefault(t *testing.T) {
	tempDir := t.TempDir()

	configPath := filepath.Join(tempDir, "config.json")

	t.Run("Missing config file", func(t *testing.T) {
		_, message, err := readConfigurationOrCreateDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error creating new config: %v", err)
		}

		if message == nil {
			t.Error("expected to get a user message about the default configuration being created but got nil")
		}

		config, message, err := readConfigurationOrCreateDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error reading newly created default config: %v", err)
		}

		if message != nil {
			t.Errorf("unexpected user message when reading newly created default configuration: %v", *message)
		}

		if config.Version != currentConfigurationVersion {
			t.Errorf("unexpected version %d for newly created default configuration", config.Version)
		}
	})

	t.Run("Legacy config file", func(t *testing.T) {
		writeTestFile(t, configPath, testLegacyConfig)

		config, message, err := readConfigurationOrCreateDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}

		if message == nil {
			t.Error("expected to get a user message about the configuration being migrated but got nil")
		}

		if config.Version != currentConfigurationVersion {
			t.Errorf("migrated configuration should have the latest version but it has %d", config.Version)
		}
	})

	t.Run("Future config file", func(t *testing.T) {
		writeTestFile(t, configPath, testFutureConfig)

		_, message, err := readConfigurationOrCreateDefault(configPath)
		if err == nil {
			t.Error("expected to get an error reading invalid (future version) configuration")
		}

		if message != nil {
			t.Errorf("unexpected user message reading invalid (future version) configuration: %v", *message)
		}
	})

	t.Run("Normal config file", func(t *testing.T) {
		writeTestFile(t, configPath, testConfig)

		config, message, err := readConfigurationOrCreateDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}

		if message != nil {
			t.Errorf("unexpected user message reading configuration: %v", *message)
		}

		if len(config.Tools) != 1 {
			t.Errorf("expected exactly one tool in example configuration but got %d", len(config.Tools))
		}
		if config.GitHubToken != "configured-token" {
			t.Errorf("got GitHub token %q, expected %q", config.GitHubToken, "configured-token")
		}
	})

	t.Run("Config file with duplicates", func(t *testing.T) {
		writeTestFile(t, configPath, testInvalidConfig)

		_, _, err := readConfigurationOrCreateDefault(configPath)
		if err == nil {
			t.Fatalf("expected to get an error when reading invalid config")
		}
	})
}

func TestValidateTargets(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		firstName  string
		secondName string
	}{
		{
			name:       "Names differing only by case",
			goos:       "linux",
			firstName:  "Foo",
			secondName: "foo",
		},
		{
			name:       "Implicit Windows executable suffix",
			goos:       "windows",
			firstName:  "foo",
			secondName: "foo.exe",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validTestConfig()
			config.Tools = map[string]Tool{
				"first":  validTestTool(test.firstName),
				"second": validTestTool(test.secondName),
			}

			if err := config.validate(test.goos); err == nil {
				t.Errorf("expected %q and %q to conflict on %s", test.firstName, test.secondName, test.goos)
			}
		})
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Configuration)
	}{
		{
			name: "Empty installation directory",
			modify: func(config *Configuration) {
				config.InstallationDirectory = ""
			},
		},
		{
			name: "Empty owner",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Owner = ""
				config.Tools["tool"] = tool
			},
		},
		{
			name: "Empty repository",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Repository = ""
				config.Tools["tool"] = tool
			},
		},
		{
			name: "Empty asset pattern",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Asset.Pattern = ""
				config.Tools["tool"] = tool
			},
		},
		{
			name: "Missing asset regex",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Asset.Regex = nil
				config.Tools["tool"] = tool
			},
		},
		{
			name: "Mismatched asset regex",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Asset.Regex = regexp.MustCompile(`windows\.zip$`)
				config.Tools["tool"] = tool
			},
		},
		{
			name: "No binaries",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Binaries = nil
				config.Tools["tool"] = tool
			},
		},
		{
			name: "Invalid source name",
			modify: func(config *Configuration) {
				tool := config.Tools["tool"]
				tool.Binaries[0].SourceNames = []string{"../tool"}
				config.Tools["tool"] = tool
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validTestConfig()
			test.modify(&config)

			if err := config.validate("linux"); err == nil {
				t.Error("expected validation to fail")
			}
		})
	}
}

func TestSaveValidation(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := "existing configuration"
	writeTestFile(t, configPath, original)

	config := validTestConfig()
	config.Tools["second"] = validTestTool("TOOL")

	if err := config.save(configPath, false); err == nil {
		t.Fatal("expected saving an invalid configuration to fail")
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read existing configuration: %v", err)
	}
	if string(contents) != original {
		t.Errorf("existing configuration was altered: got %q, expected %q", contents, original)
	}
}
