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

func readConfiguration(path string) (Configuration, error) {
	bytes, err := os.ReadFile(replaceTildePath(path))
	if err != nil {
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

const defaultConfiguration = `{
	"install_dir": "~/.local/bin",
	"tools": {
		"bat": {
			"binaries": [
				{
					"name": "bat",
					"rename_to": ""
				}
			],
			"owner": "sharkdp",
			"repository": "bat",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Better cat"
		},
		"delta": {
			"binaries": [
				{
					"name": "delta",
					"rename_to": ""
				}
			],
			"owner": "dandavison",
			"repository": "delta",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Diff tool"
		},
		"eza": {
			"binaries": [
				{
					"name": "eza",
					"rename_to": ""
				}
			],
			"owner": "eza-community",
			"repository": "eza",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-gnu\\.zip$",
			"description": "Better ls (replacement of exa which is unmaintained)"
		},
		"fd": {
			"binaries": [
				{
					"name": "fd",
					"rename_to": ""
				}
			],
			"owner": "sharkdp",
			"repository": "fd",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Better find"
		},
		"hyperfine": {
			"binaries": [
				{
					"name": "hyperfine",
					"rename_to": ""
				}
			],
			"owner": "sharkdp",
			"repository": "hyperfine",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Benchmark tool"
		},
		"micro": {
			"binaries": [
				{
					"name": "micro",
					"rename_to": ""
				}
			],
			"owner": "zyedidia",
			"repository": "micro",
			"linux_asset": "linux64\\.tar\\.gz$",
			"windows_asset": "win64\\.zip$",
			"description": "Command-line editor"
		},
		"ripgrep": {
			"binaries": [
				{
					"name": "rg",
					"rename_to": ""
				}
			],
			"owner": "burntsushi",
			"repository": "ripgrep",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Better grep"
		},
		"sd": {
			"binaries": [
				{
					"name": "sd",
					"rename_to": ""
				}
			],
			"owner": "chmln",
			"repository": "sd",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Better sed"
		},
		"starship": {
			"binaries": [
				{
					"name": "starship",
					"rename_to": ""
				}
			],
			"owner": "starship",
			"repository": "starship",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "Cross-shell custom prompt"
		},
		"tealdeer": {
			"binaries": [
				{
					"name": "tealdeer",
					"rename_to": "tldr"
				}
			],
			"owner": "dbrgn",
			"repository": "tealdeer",
			"linux_asset": "tealdeer-linux-x86_64-musl$",
			"windows_asset": "windows-x86_64-msvc.exe$",
			"description": "Command-line cheatsheets"
		},
		"tokei": {
			"binaries": [
				{
					"name": "tokei",
					"rename_to": ""
				}
			],
			"owner": "XAMPPRocky",
			"repository": "tokei",
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc.exe$",
			"description": "Code line counting tool"
		}
	}
}`

func writeDefaultConfiguration(path string) error {
	tmp := []byte(defaultConfiguration)
	defaultConfig, err := parseConfiguration(tmp)
	if err != nil {
		return fmt.Errorf("failed to parse default configuration: %w", err)
	}

	err = defaultConfig.save(path, true)
	if err != nil {
		return fmt.Errorf("error creating configuration file: %w", err)
	}

	fmt.Printf("Created default configuration: '%s'\n", path)

	return nil
}
