// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Cache struct {
	Tools map[string]string `json:"tools"`
}

func (cache *Cache) writeCache() error {
	filePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(filePath)
	err = makeOutputDirectory(&cacheDir)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(*cache, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, bytes, 0644)
}

func getCache() (Cache, error) {
	result := Cache{Tools: make(map[string]string)}

	filePath, err := getCacheFilePath()
	if err != nil {
		return result, err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return result, nil
	} else if err != nil {
		return result, err
	}

	bytes, err := os.ReadFile(replaceTildePath(filePath))
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
