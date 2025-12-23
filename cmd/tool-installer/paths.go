// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const appName = "tool-installer"
const cacheFileName = "tool-versions.json"
const configFileName = "config.json"

func addExeSuffix(fileName string) string {
	if !strings.HasSuffix(fileName, ".exe") {
		return fileName + ".exe"
	}

	return fileName
}

func getCacheFilePath() (string, error) {
	if cacheDir := os.Getenv("TOOLI_CACHE_DIRECTORY"); cacheDir != "" {
		return filepath.Clean(filepath.Join(cacheDir, cacheFileName)), nil
	}

	baseDir := ""

	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		baseDir = xdgCacheHome
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}

		baseDir = cacheDir
	}

	return filepath.Clean(filepath.Join(baseDir, appName, cacheFileName)), nil
}

func getConfigFilePath() (string, error) {
	if configDir := os.Getenv("TOOLI_CONFIG_DIRECTORY"); configDir != "" {
		return filepath.Clean(filepath.Join(configDir, configFileName)), nil
	}

	baseDir := ""

	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		baseDir = xdgConfigHome
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		baseDir = configDir
	}

	return filepath.Clean(filepath.Join(baseDir, appName, configFileName)), nil
}

func makeOutputDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory ('%s'): %w", path, err)
	}

	return nil
}
