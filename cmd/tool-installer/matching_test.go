// SPDX-License-Identifier: Apache-2.0

package main

import (
	"slices"
	"testing"
)

var testAssets = []Asset{
	{Name: "standard-x86_64-pc-windows-msvc.zip"},
	{Name: "standard-x86_64-pc-windows-gnu.zip"},
	{Name: "standard-x86_64-unknown-linux-gnu.tar.gz"},
	{Name: "standard-aarch64-apple-darwin.tar.gz"},
}

func TestMatchBestAssetName(t *testing.T) {
	tests := []struct {
		name               string
		goos               string
		goarch             string
		expectedAssetNames []string
	}{
		{"Windows (x64)", "windows", "amd64", []string{"standard-x86_64-pc-windows-msvc.zip"}},
		{"Linux (x64)", "linux", "amd64", []string{"standard-x86_64-unknown-linux-gnu.tar.gz"}},
		{"Apple (ARM)", "darwin", "arm64", []string{"standard-aarch64-apple-darwin.tar.gz"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidates := matchBestAssetName(testAssets, test.goos, test.goarch)

			names := make([]string, len(candidates))
			for i, c := range candidates {
				names[i] = c.asset.Name
			}

			slices.Sort(names)
			slices.Sort(test.expectedAssetNames)

			if !slices.Equal(names, test.expectedAssetNames) {
				t.Errorf("matched assets are wrong: got %v, expected %v", names, test.expectedAssetNames)
			}
		})
	}
}
