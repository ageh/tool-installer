// SPDX-License-Identifier: Apache-2.0

package main

import (
	"slices"
	"strings"
	"testing"
)

func TestKnownToolsPlatformNames(t *testing.T) {
	allowedPlatforms := []string{"linux/amd64", "linux/arm64", "windows/amd64", "windows/arm64", "darwin/amd64", "darwin/arm64"}

	for name, tool := range knownTools {
		for platform := range tool.AssetNames {
			if !slices.Contains(allowedPlatforms, platform) {
				t.Errorf("invalid platform name entry %q in known tool %q", platform, name)
			}
		}
	}
}

func TestKnownToolsBinaryPlatforms(t *testing.T) {
	for name, tool := range knownTools {
		for _, binary := range tool.Binaries {
			for _, platform := range binary.Platforms {
				if !isValidPlatformEntry(platform) {
					t.Errorf("invalid platform entry %q for binary %q in known tool %q", platform, binary.Name, name)
				}
			}
		}
	}
}

func TestFilterBinariesForPlatform(t *testing.T) {
	binaries := []Binary{
		{Name: "universal"},
		{Name: "linux-only", Platforms: []string{"linux"}},
		{Name: "windows-amd64-only", Platforms: []string{"windows/amd64"}},
	}

	result := filterBinariesForPlatform(binaries, "windows", "amd64")

	if len(result) != 2 {
		t.Fatalf("expected 2 binaries, got %d: %v", len(result), result)
	}

	if result[0].Name != "universal" || result[1].Name != "windows-amd64-only" {
		t.Errorf("unexpected filtered binaries: %v", result)
	}
}

func TestKnownToolsBinaryNames(t *testing.T) {
	seen := make(map[string]int)

	for _, tool := range knownTools {
		for _, binary := range tool.Binaries {
			seen[strings.ToLower(binary.Name)]++
		}
	}

	for name, count := range seen {
		if count > 1 {
			t.Errorf("unexpected duplicate binary name found in known tools: %q: %v", name, count)
		}
	}
}
