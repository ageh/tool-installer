// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Binary struct {
	Name     string `json:"name"`
	RenameTo string `json:"rename_to"`
}

func (binary *Binary) MarshalJSON() ([]byte, error) {
	if binary == nil {
		return []byte("null"), nil
	}

	return json.Marshal(&struct {
		Name     string `json:"name"`
		RenameTo string `json:"rename_to"`
	}{
		Name:     strings.TrimSuffix(binary.Name, ".exe"),
		RenameTo: strings.TrimSuffix(binary.RenameTo, ".exe"),
	})
}

type Tool struct {
	Binaries     []Binary `json:"binaries"`
	Owner        string   `json:"owner"`
	Repository   string   `json:"repository"`
	LinuxAsset   string   `json:"linux_asset"`
	WindowsAsset string   `json:"windows_asset"`
	Description  string   `json:"description"`
}

type Configuration struct {
	InstallationDirectory string          `json:"install_dir"`
	Tools                 map[string]Tool `json:"tools"`
}

func parseConfiguration(input []byte) (Configuration, error) {
	var config Configuration

	err := json.Unmarshal(input, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if runtime.GOOS == "windows" {
		for name, tool := range config.Tools {
			_, err := regexp.Compile(tool.WindowsAsset)
			if err != nil {
				return config, fmt.Errorf("error in Windows asset regex for tool '%s': %w", name, err)
			}
			_, err = regexp.Compile(tool.LinuxAsset)
			if err != nil {
				return config, fmt.Errorf("error in Linux asset regex for tool '%s': %w", name, err)
			}

			for i, b := range tool.Binaries {
				config.Tools[name].Binaries[i].Name = addExeSuffix(b.Name)
				if b.RenameTo != "" {
					config.Tools[name].Binaries[i].RenameTo = addExeSuffix(b.RenameTo)
				}
			}
		}
	}

	return config, nil
}

func readConfigurationOrCreateDefault(path string) (Configuration, error) {
	bytes, err := os.ReadFile(replaceTildePath(path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			config := getDefaultConfiguration()
			err := config.save(path, false)
			if err != nil {
				return Configuration{}, fmt.Errorf("failed to write default configuration to disk: %w", err)
			}

			return config, nil
		}

		return Configuration{}, err
	}

	return parseConfiguration(bytes)
}

func (config *Configuration) save(path string, promptOverride bool) error {
	filePath := replaceTildePath(path)
	dirName := filepath.Dir(filePath)

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create the directory for configuration writing: %w", err)
	}

	_, err = os.Stat(filePath)
	if err == nil {
		if promptOverride {
			fmt.Print("A file already exists at that location. Overwrite? [y/N]")
			var input string
			_, err := fmt.Scan(&input)
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			if input != "y" && input != "Y" {
				return nil
			}
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error when checking if target file already exists: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating configuration file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("error writing configuration to file: %w", err)
	}

	return nil
}

var defaultTools = []string{
	"bat",
	"delta",
	"eza",
	"fd",
	"hyperfine",
	"micro",
	"ripgrep",
	"ruff",
	"sd",
	"starship",
	"tealdeer",
	"tokei",
	"ty",
	"uv",
}

func getDefaultConfiguration() Configuration {
	tools := make(map[string]Tool)
	for _, name := range defaultTools {
		tool, found := knownTools[name]
		if !found {
			panic(fmt.Sprintf("Could not find default tool '%s' in known tools", name))
		}

		tools[name] = tool
	}

	return Configuration{InstallationDirectory: "~/.local/bin", Tools: tools}
}
