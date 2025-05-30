// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	Tools map[string]string `json:"tools"`
}

func (cache *Cache) add(tool string, version string) {
	cache.Tools[tool] = version
}

func (cache *Cache) remove(tool string) {
	delete(cache.Tools, tool)
}

func (cache *Cache) writeCache() error {
	errMessage := "error writing to cache: %w"

	filePath, err := getCacheFilePath()
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	cacheDir := filepath.Dir(filePath)
	err = makeOutputDirectory(cacheDir)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	bytes, err := json.MarshalIndent(*cache, "", "\t")
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	err = os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

func getCache() (Cache, error) {
	result := Cache{Tools: make(map[string]string)}

	errMessage := "error obtaining cache: %w"

	filePath, err := getCacheFilePath()
	if err != nil {
		return result, fmt.Errorf(errMessage, err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return result, nil
	} else if err != nil {
		return result, fmt.Errorf(errMessage, err)
	}

	bytes, err := os.ReadFile(replaceTildePath(filePath))
	if err != nil {
		return result, fmt.Errorf(errMessage, err)
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, fmt.Errorf(errMessage, err)
	}

	return result, nil
}
