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
