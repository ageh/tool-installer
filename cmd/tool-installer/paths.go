// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func addExeSuffix(fileName string) string {
	if !strings.HasSuffix(fileName, ".exe") {
		return fileName + ".exe"
	}

	return fileName
}

func getCacheFilePath() (string, error) {
	baseDir := ""

	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		baseDir = xdgCacheHome
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		baseDir = filepath.Join(usr.HomeDir, ".cache")
	}

	return filepath.Join(baseDir, "tool-installer", "tool-versions.json"), nil
}

func getConfigFilePath() (string, error) {
	baseDir := ""

	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		baseDir = xdgConfigHome
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		baseDir = filepath.Join(usr.HomeDir, ".config")
	}

	return filepath.Join(baseDir, "tool-installer", "config.json"), nil
}

func replaceTildePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		return dir
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, path[2:])
	} else {
		return path
	}
}

func makeOutputDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory ('%s'): %w", path, err)
	}

	return nil
}
