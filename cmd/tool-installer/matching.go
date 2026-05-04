// SPDX-License-Identifier: Apache-2.0

package main

import (
	"regexp"
)

var typicalAssets = map[string]map[string][]AssetRegex{
	"windows": {
		"amd64": {
			{Pattern: standardAssetWindowsx64, Regex: regexp.MustCompile(standardAssetWindowsx64)},
			{Pattern: standardAssetWindowsGnux64, Regex: regexp.MustCompile(standardAssetWindowsGnux64)},
		},
		"arm64": {
			{Pattern: standardAssetWindowsArm, Regex: regexp.MustCompile(standardAssetWindowsArm)},
		},
	},
	"linux": {
		"amd64": {
			{Pattern: standardAssetLinuxx64, Regex: regexp.MustCompile(standardAssetLinuxx64)},
			{Pattern: standardAssetLinuxMuslx64, Regex: regexp.MustCompile(standardAssetLinuxMuslx64)},
		},
		"arm64": {
			{Pattern: standardAssetLinuxArm, Regex: regexp.MustCompile(standardAssetLinuxArm)},
			{Pattern: standardAssetLinuxMuslArm, Regex: regexp.MustCompile(standardAssetLinuxMuslArm)},
		},
	},
	"darwin": {
		"amd64": {
			{Pattern: standardAssetApplex64, Regex: regexp.MustCompile(standardAssetApplex64)},
		},
		"arm64": {
			{Pattern: standardAssetAppleArm, Regex: regexp.MustCompile(standardAssetAppleArm)},
		},
	},
}

type AssetCandidate struct {
	asset      Asset
	assetRegex AssetRegex
}

func matchBestAssetName(assets []Asset, goos string, goarch string) []AssetCandidate {
	results := make([]AssetCandidate, 0)

	patterns := typicalAssets[goos][goarch]

	for _, pattern := range patterns {
		for _, asset := range assets {
			if pattern.Regex.MatchString(asset.Name) {
				results = append(results, AssetCandidate{asset: asset, assetRegex: pattern})
			}
		}

		if len(results) != 0 {
			return results
		}
	}

	return results
}
