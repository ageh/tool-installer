// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const appName = "tool-installer"
const cacheFileName = "tool-versions.json"
const configFileName = "config.json"

func addExeSuffix(fileName string) string {
	if fileName == "" {
		return fileName
	}

	if !strings.HasSuffix(strings.ToLower(fileName), ".exe") {
		return fileName + ".exe"
	}

	return fileName
}

func addExeSuffixes(names []string) []string {
	result := make([]string, len(names))

	for i, n := range names {
		result[i] = addExeSuffix(n)
	}

	return result
}

func stripExeSuffix(fileName string) string {
	if strings.HasSuffix(strings.ToLower(fileName), ".exe") {
		return fileName[:len(fileName)-len(".exe")]
	}

	return fileName
}

func stripExeSuffixes(names []string) []string {
	result := make([]string, len(names))

	for i, n := range names {
		result[i] = stripExeSuffix(n)
	}

	return result
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
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return fmt.Errorf("error creating output directory ('%s'): %w", path, err)
	}

	return nil
}

var forbiddenChars = regexp.MustCompile(`[\<\>\:\"\/\\\|\?\*\x00-\x1F]`)
var reservedWindowsNames = regexp.MustCompile(`^(?i)(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\..*)?$`)

func isPlainFilename(filename string) bool {
	if len(strings.TrimSpace(filename)) == 0 || strings.ContainsRune(filename, 0) {
		return false
	}

	if strings.HasPrefix(filename, ".") {
		return false
	}

	if forbiddenChars.MatchString(filename) {
		return false
	}

	if reservedWindowsNames.MatchString(filename) {
		return false
	}

	cleaned := filepath.Clean(filename)
	if filepath.Base(cleaned) != filename {
		return false
	}

	return true
}
