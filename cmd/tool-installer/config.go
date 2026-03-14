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

const currentConfigurationVersion = 2

type Binary struct {
	Name     string `json:"-"`
	RenameTo string `json:"-"`
}

func (binary Binary) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name     string `json:"name"`
		RenameTo string `json:"rename_to,omitempty"`
	}{
		Name:     strings.TrimSuffix(binary.Name, ".exe"),
		RenameTo: strings.TrimSuffix(binary.RenameTo, ".exe"),
	})
}

func (b *Binary) UnmarshalJSON(bytes []byte) error {
	var result struct {
		Name     string `json:"name"`
		RenameTo string `json:"rename_to,omitempty"`
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("failed to parse JSON into Binary type: %w", err)
	}

	b.Name = result.Name

	if result.RenameTo == "" {
		return nil
	}

	baseName := filepath.Base(result.RenameTo)
	if baseName == "." || baseName == ".." || strings.ContainsAny(baseName, `/\`) {
		return fmt.Errorf("invalid rename_to ('%s'): must be a plain filename", b.RenameTo)
	}

	b.RenameTo = result.RenameTo

	return nil
}

type AssetRegex struct {
	Pattern string         `json:"-"`
	Regex   *regexp.Regexp `json:"-"`
}

func (a AssetRegex) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Pattern)
}

func (a *AssetRegex) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return fmt.Errorf("regex must be a string: %w", err)
	}

	re, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", s, err)
	}

	a.Pattern = s
	a.Regex = re

	return nil
}

func stringToAssetRegex(input string) (AssetRegex, error) {
	re, err := regexp.Compile(input)
	if err != nil {
		return AssetRegex{}, err
	}

	return AssetRegex{Pattern: input, Regex: re}, nil
}

type Tool struct {
	Binaries    []Binary   `json:"binaries"`
	Owner       string     `json:"owner"`
	Repository  string     `json:"repository"`
	Asset       AssetRegex `json:"asset"`
	Description string     `json:"description"`
}

type Configuration struct {
	Version               int             `json:"version"`
	InstallationDirectory string          `json:"install_dir"`
	Tools                 map[string]Tool `json:"tools"`
}

func (c *Configuration) getSanitizedInstallationDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if c.InstallationDirectory == "~" {
		return filepath.Clean(homeDir), nil
	} else if strings.HasPrefix(c.InstallationDirectory, "~/") {
		return filepath.Clean(filepath.Join(homeDir, c.InstallationDirectory[2:])), nil
	} else {
		return filepath.Clean(c.InstallationDirectory), nil
	}
}

func (config *Configuration) save(path string, promptOverride bool) error {
	dirName := filepath.Dir(path)

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create the directory for configuration writing: %w", err)
	}

	_, err = os.Stat(path)
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

	file, err := os.Create(path)
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

func parseConfiguration(input []byte) (Configuration, error) {
	var config Configuration

	err := json.Unmarshal(input, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if runtime.GOOS == "windows" {
		for name, tool := range config.Tools {
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

func getConfigurationVersion(input []byte) (int, error) {
	var versionTester struct {
		Version int `json:"version"`
	}

	if err := json.Unmarshal(input, &versionTester); err != nil {
		return -1, err
	}

	return versionTester.Version, nil
}

func readConfigurationOrCreateDefault(path string) (Configuration, *UserMessage, error) {
	var message *UserMessage = nil
	bytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			config, err := getDefaultConfiguration()
			message = &UserMessage{Type: Info, Tool: "tooli", Content: "Default configuration has been created because no configuration file existed yet"}
			if err != nil {
				return Configuration{}, message, err
			}

			err = config.save(path, false)
			if err != nil {
				return Configuration{}, message, fmt.Errorf("failed to write default configuration to disk: %w", err)
			}

			return config, message, nil
		}

		return Configuration{}, message, err
	}

	version, err := getConfigurationVersion(bytes)
	if err != nil {
		return Configuration{}, message, fmt.Errorf("failed to obtain configuration version: %w", err)
	}

	if version > currentConfigurationVersion {
		return Configuration{}, message, fmt.Errorf("invalid version (%d), the latest supported configuration version is %d", version, currentConfigurationVersion)
	}

	if version < currentConfigurationVersion {
		config, err := migrateConfiguration(bytes)
		if err != nil {
			return Configuration{}, message, err
		}

		message = &UserMessage{Type: Info, Tool: "tooli", Content: "Configuration has been automatically migrated from the old format, please check it"}

		err = config.save(path, false)
		if err != nil {
			return Configuration{}, message, fmt.Errorf("could not save automatically migrated configuration: %w", err)
		}

		return config, message, nil
	}

	config, err := parseConfiguration(bytes)

	return config, message, err
}

type ToolV1 struct {
	Binaries     []Binary `json:"binaries"`
	Owner        string   `json:"owner"`
	Repository   string   `json:"repository"`
	LinuxAsset   string   `json:"linux_asset"`
	WindowsAsset string   `json:"windows_asset"`
	Description  string   `json:"description"`
}

type ConfigurationV1 struct {
	InstallationDirectory string            `json:"install_dir"`
	Tools                 map[string]ToolV1 `json:"tools"`
}

func migrateConfiguration(input []byte) (Configuration, error) {
	var oldConfig ConfigurationV1

	err := json.Unmarshal(input, &oldConfig)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to parse old configuration: %w", err)
	}

	var result = Configuration{
		Version:               currentConfigurationVersion,
		InstallationDirectory: oldConfig.InstallationDirectory,
		Tools:                 make(map[string]Tool),
	}

	switch runtime.GOOS {
	case "windows":
		for name, tool := range oldConfig.Tools {
			re, err := regexp.Compile(tool.WindowsAsset)
			if err != nil {
				return Configuration{}, fmt.Errorf("invalid Windows asset in old configuration for tool '%s': %w", name, err)
			}

			result.Tools[name] = Tool{
				Binaries:    tool.Binaries,
				Owner:       tool.Owner,
				Repository:  tool.Repository,
				Asset:       AssetRegex{Pattern: tool.WindowsAsset, Regex: re},
				Description: tool.Description,
			}
		}
	case "linux":
		for name, tool := range oldConfig.Tools {
			re, err := regexp.Compile(tool.LinuxAsset)
			if err != nil {
				return Configuration{}, fmt.Errorf("invalid Linux asset in old configuration for tool '%s': %w", name, err)
			}

			result.Tools[name] = Tool{
				Binaries:    tool.Binaries,
				Owner:       tool.Owner,
				Repository:  tool.Repository,
				Asset:       AssetRegex{Pattern: tool.LinuxAsset, Regex: re},
				Description: tool.Description,
			}
		}
	default:
		return Configuration{}, errors.New("failed to convert old configuration, this platform was not supported in the old format")
	}

	return result, nil
}

var defaultTools = []string{
	"bat",
	"bun",
	"delta",
	"deno",
	"eza",
	"fd",
	"hyperfine",
	"micro",
	"pandoc",
	"ripgrep",
	"ruff",
	"sd",
	"starship",
	"tealdeer",
	"tokei",
	"ty",
	"typst",
	"uv",
}

func getDefaultConfiguration() (Configuration, error) {
	tools := make(map[string]Tool)
	for _, name := range defaultTools {
		tool, found := knownTools[name]
		if !found {
			return Configuration{}, fmt.Errorf("could not find default tool '%s' in known tools", name)
		}

		tmp, err := tool.intoToolForPlatform()
		if err != nil {
			return Configuration{}, fmt.Errorf("failed to obtain tool from known tools: %w", err)
		}

		tools[name] = tmp
	}

	return Configuration{InstallationDirectory: "~/.local/bin", Tools: tools}, nil
}
