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
	"slices"
	"strings"
)

const currentConfigurationVersion = 3

type Binary struct {
	Name        string   `json:"name"`
	RenameTo    string   `json:"rename_to,omitempty"` // Deprecated, only kept for migration
	SourceNames []string `json:"source_names,omitempty"`
}

func (binary Binary) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name        string   `json:"name"`
		SourceNames []string `json:"source_names,omitempty"`
	}{
		Name:        stripExeSuffix(binary.Name),
		SourceNames: stripExeSuffixes(binary.SourceNames),
	})
}

func (b *Binary) UnmarshalJSON(bytes []byte) error {
	var result struct {
		Name        string   `json:"name"`
		RenameTo    string   `json:"rename_to,omitempty"`
		SourceNames []string `json:"source_names,omitempty"`
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("failed to parse JSON into Binary type: %w", err)
	}

	if !isPlainFilename(result.Name) {
		return fmt.Errorf("invalid name ('%s'): must be a plain filename", result.Name)
	}

	b.Name = result.Name
	b.SourceNames = result.SourceNames

	if result.RenameTo == "" {
		return nil
	}

	if !isPlainFilename(result.RenameTo) {
		return fmt.Errorf("invalid rename_to ('%s'): must be a plain filename", result.RenameTo)
	}

	b.RenameTo = result.RenameTo

	return nil
}

func (binary Binary) getTargetName() string {
	return binary.Name
}

func (b Binary) getTargetKey(goos string) string {
	name := b.getTargetName()

	if goos == "windows" {
		name = addExeSuffix(name)
	}

	return strings.ToLower(name)
}

func (binary Binary) hasSourceName(name string) bool {
	return name == binary.Name || slices.Contains(binary.SourceNames, name)
}

func migrateBinary(binary Binary) Binary {
	var result Binary

	if binary.RenameTo != "" {
		result.Name = binary.RenameTo
		result.SourceNames = append(result.SourceNames, binary.Name)
	} else {
		result.Name = binary.Name
	}

	return result
}

func migrateBinaries(binaries []Binary) []Binary {
	result := make([]Binary, len(binaries))

	for i, binary := range binaries {
		result[i] = migrateBinary(binary)
	}

	return result
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

	if s == "" {
		return errors.New("invalid regex pattern: must not be empty")
	}

	re, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", s, err)
	}

	a.Pattern = s
	a.Regex = re

	return nil
}

type Tool struct {
	Binaries    []Binary   `json:"binaries"`
	Owner       string     `json:"owner"`
	Repository  string     `json:"repository"`
	Asset       AssetRegex `json:"asset"`
	Description string     `json:"description"`
}

func (t Tool) forCurrentPlatform(goos string) Tool {
	if goos != "windows" {
		return t
	}

	result := t
	result.Binaries = make([]Binary, len(t.Binaries))

	for i, binary := range t.Binaries {
		result.Binaries[i] = Binary{
			Name:        addExeSuffix(binary.Name),
			SourceNames: addExeSuffixes(binary.SourceNames),
		}
	}

	return result
}

type Configuration struct {
	Version               int             `json:"version"`
	InstallationDirectory string          `json:"install_dir"`
	GitHubToken           string          `json:"github_token,omitempty"`
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
	err := config.validate(runtime.GOOS)
	if err != nil {
		return err
	}

	dirName := filepath.Dir(path)

	err = os.MkdirAll(dirName, 0o755)
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

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("error creating configuration file: %w", err)
	}
	defer file.Close()

	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("error setting configuration file permissions: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("error writing configuration to file: %w", err)
	}

	return nil
}

func (config *Configuration) validate(goos string) error {
	if config.InstallationDirectory == "" {
		return errors.New("invalid configuration: installation directory must be non-empty")
	}

	seen := make(map[string]string)
	for toolName, tool := range config.Tools {
		if tool.Owner == "" {
			return errors.New("invalid configuration: owner must be non-empty")
		}

		if tool.Repository == "" {
			return errors.New("invalid configuration: repository must be non-empty")
		}

		if tool.Asset.Pattern == "" {
			return fmt.Errorf("invalid configuration: tool %q has no asset pattern", toolName)
		}

		if tool.Asset.Regex == nil {
			return fmt.Errorf("invalid configuration: tool %q has no compiled asset regex", toolName)
		}

		if tool.Asset.Regex.String() != tool.Asset.Pattern {
			return fmt.Errorf("invalid configuration: asset pattern and compiled regex differ for tool %q", toolName)
		}

		if len(tool.Binaries) == 0 {
			return fmt.Errorf("invalid configuration: tool %q has no configured binaries", toolName)
		}

		for _, binary := range tool.Binaries {
			key := binary.getTargetKey(goos)

			if previous, found := seen[key]; found {
				return fmt.Errorf("invalid configuration: binary %q from tool %q conflicts with tool %q", binary.Name, toolName, previous)
			}

			seen[key] = toolName

			for _, sourceName := range binary.SourceNames {
				if !isPlainFilename(sourceName) {
					return fmt.Errorf("invalid configuration: tool %q has an invalid source name %q", toolName, sourceName)
				}
			}
		}
	}

	return nil
}

func parseConfiguration(input []byte) (Configuration, error) {
	var config Configuration

	err := json.Unmarshal(input, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return config, nil
}

func normalizeConfiguration(config Configuration, goos string) Configuration {
	result := config
	result.Tools = make(map[string]Tool, len(config.Tools))

	for name, tool := range config.Tools {
		result.Tools[name] = tool.forCurrentPlatform(goos)
	}

	return result
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
	var message *UserMessage
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

			normalized := normalizeConfiguration(config, runtime.GOOS)
			if err = normalized.validate(runtime.GOOS); err != nil {
				return Configuration{}, message, err
			}

			return normalized, message, nil
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
		config, err := migrateConfiguration(bytes, version)
		if err != nil {
			return Configuration{}, message, err
		}

		if config.Version != currentConfigurationVersion {
			return Configuration{}, message, fmt.Errorf("incomplete configuration migration: expected version %d but only migrated to version %d", currentConfigurationVersion, config.Version)
		}

		message = &UserMessage{Type: Info, Tool: "tooli", Content: fmt.Sprintf("Configuration has been automatically migrated from the old format, please check it (%d -> %d)", version, currentConfigurationVersion)}

		err = config.save(path, false)
		if err != nil {
			return Configuration{}, message, fmt.Errorf("could not save automatically migrated configuration: %w", err)
		}

		normalized := normalizeConfiguration(config, runtime.GOOS)
		if err = normalized.validate(runtime.GOOS); err != nil {
			return Configuration{}, message, err
		}

		return normalized, message, nil
	}

	config, err := parseConfiguration(bytes)
	if err != nil {
		return Configuration{}, message, err
	}

	normalized := normalizeConfiguration(config, runtime.GOOS)
	if err = normalized.validate(runtime.GOOS); err != nil {
		return Configuration{}, message, err
	}

	return normalized, message, nil
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

func migrateConfigV1toV3(input []byte) (Configuration, error) {
	var oldConfig ConfigurationV1

	err := json.Unmarshal(input, &oldConfig)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to parse old configuration: %w", err)
	}

	var result = Configuration{
		Version:               3,
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
				Binaries:    migrateBinaries(tool.Binaries),
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
				Binaries:    migrateBinaries(tool.Binaries),
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

func migrateConfigV2toV3(input []byte) (Configuration, error) {
	config, err := parseConfiguration(input)
	if err != nil {
		return Configuration{}, err
	}

	config.Version = 3
	for name, tool := range config.Tools {
		tool.Binaries = migrateBinaries(tool.Binaries)
		config.Tools[name] = tool
	}

	return config, nil
}

func migrateConfiguration(input []byte, oldVersion int) (Configuration, error) {
	switch oldVersion {
	case 0, 1:
		return migrateConfigV1toV3(input)
	case 2:
		return migrateConfigV2toV3(input)
	default:
		return Configuration{}, fmt.Errorf("invalid version %d passed to migrateConfiguration", oldVersion)
	}
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
			if errors.Is(err, ErrUnsupportedPlatform) {
				continue
			}

			return Configuration{}, fmt.Errorf("failed to obtain tool from known tools: %w", err)
		}

		tools[name] = tmp
	}

	return Configuration{Version: currentConfigurationVersion, InstallationDirectory: "~/.local/bin", Tools: tools}, nil
}
