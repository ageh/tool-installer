// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"testing"
)

func TestBinaryMarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    Binary
		expected string
	}{
		{name: "No rename_to", input: Binary{Name: "rg"}, expected: `{"name":"rg"}`},
		{name: "Empty rename_to", input: Binary{Name: "fd", RenameTo: ""}, expected: `{"name":"fd"}`},
		{name: "With rename_to", input: Binary{Name: "fd.exe", RenameTo: "fd.exe"}, expected: `{"name":"fd","rename_to":"fd"}`},
		{name: "Windows exe suffix stripped", input: Binary{Name: "rg.exe"}, expected: `{"name":"rg"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tc.expected {
				t.Errorf("got %q, expected %q", got, tc.expected)
			}
		})
	}
}

func TestBinaryUnmarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`not json`), &b); err == nil {
			t.Error("expected an error for invalid JSON, got nil")
		}
	})

	var tests = []struct {
		name     string
		input    string
		expected Binary
	}{
		{"No rename_to", `{"name": "rg"}`, Binary{Name: "rg"}},
		{"Empty rename_to", `{"name": "rg", "rename_to": ""}`, Binary{Name: "rg"}},
		{"Non-empty rename_to", `{"name": "rg", "rename_to": "ripgrep"}`, Binary{Name: "rg", RenameTo: "ripgrep"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var b Binary
			if err := json.Unmarshal([]byte(test.input), &b); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if b.Name != test.expected.Name {
				t.Errorf("got name %q but expected %q", b.Name, test.expected.Name)
			}

			if b.RenameTo != test.expected.RenameTo {
				t.Errorf("got rename_to %q but expected %q", b.RenameTo, test.expected.RenameTo)
			}
		})
	}

	t.Run("invalid rename_to: contains period", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "."}`), &b); err == nil {
			t.Error("expected an error for rename_to, got nil")
		}
	})

	t.Run("invalid rename_to: path separator", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "rip/grep"}`), &b); err == nil {
			t.Error("expected an error for rename_to, got nil")
		}
	})

	t.Run("invalid rename_to: Windows path separator", func(t *testing.T) {
		var b Binary
		if err := json.Unmarshal([]byte(`{"name": "rg", "rename_to": "rip\grep"}`), &b); err == nil {
			t.Error("expected an error for rename_to, got nil")
		}
	})
}

func TestBinaryMarshalUnmarshalJSON(t *testing.T) {
	cases := []Binary{
		{Name: "rg"},
		{Name: "tool-installer", RenameTo: "tooli"},
	}

	for _, original := range cases {
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

		if restored.RenameTo != original.RenameTo {
			t.Errorf("RenameTo mismatch: got %q, expected %q", restored.RenameTo, original.RenameTo)
		}
	}
}

func TestAssetRegexMarshalJSON(t *testing.T) {
	cases := []struct {
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tc.expected {
				t.Errorf("got %q, expected %q", got, tc.expected)
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

	var tests = []struct {
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
	tests := []string{`windows-x86_64\.tar\.gz`, `fd-v[0-9]+\.[0-9]+\.[0-9]+-x86_64-unknown-linux-musl\.tar\.gz`}

	for _, pattern := range tests {
		original := AssetRegex{Pattern: pattern}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal failed for %q: %v", pattern, err)
		}

		var restored AssetRegex
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("unmarshal failed for %q: %v", data, err)
		}

		if restored.Pattern != original.Pattern {
			t.Errorf("Pattern mismatch: got %q, expected %q", restored.Pattern, original.Pattern)
		}
		if restored.Regex == nil {
			t.Error("expected Regex to be compiled after unmarshal, got nil")
		}
	}
}

func TestDefaultConfiguration(t *testing.T) {
	config, err := getDefaultConfiguration()
	if err != nil {
		t.Errorf("getDefaultConfiguration returned an error: %v", err)
	}

	if config.Version != currentConfigurationVersion {
		t.Errorf("getDefaultConfiguration returned a wrong version number, got %d, expected %d", config.Version, currentConfigurationVersion)
	}

	if config.InstallationDirectory != "~/.local/bin" {
		t.Errorf("getDefaultConfiguration returned a wrong installation directory, got %q, expected '~/.local/bin'", config.InstallationDirectory)
	}

	for _, name := range defaultTools {
		tool, found := config.Tools[name]
		if !found {
			t.Errorf("expected to find an entry for %q in the default configuration", name)
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
