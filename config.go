// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Binary struct {
	Name     string `json:"name"`
	RenameTo string `json:"rename_to"`
}

type Tool struct {
	Binaries     []Binary `json:"binaries"`
	Owner        string   `json:"owner"`
	Repository   string   `json:"repository"`
	LinuxAsset   string   `json:"linux_asset"`
	WindowsAsset string   `json:"windows_asset"`
}

type Configuration struct {
	InstallationDirectory string          `json:"install_dir"`
	Tools                 map[string]Tool `json:"tools"`
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

func GetConfig(path string) (Configuration, error) {
	var config Configuration

	bytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}

	config.InstallationDirectory = replaceTildePath(config.InstallationDirectory)

	return config, err
}

const defaultConfiguration = `{
	"install_dir": "~/.local/bin",
	"tools": {
		"exa": {
			"binaries": [
				{
					"name": "exa",
					"rename_to": ""
				}
			],
			"owner": "ogham",
			"repository": "exa",
			"linux_asset": "linux-x86_64-musl-v0.10.1.zip",
			"windows_asset": ""
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
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
		},
		"dust": {
			"binaries": [
				{
					"name": "dust",
					"rename_to": ""
				}
			],
			"owner": "bootandy",
			"repository": "dust",
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
		},
		"fzf": {
			"binaries": [
				{
					"name": "fzf",
					"rename_to": ""
				}
			],
			"owner": "junegunn",
			"repository": "fzf",
			"linux_asset": "linux_amd64.tar.gz",
			"windows_asset": "windows_amd64.zip"
		},
		"hexyl": {
			"binaries": [
				{
					"name": "hexyl",
					"rename_to": ""
				}
			],
			"owner": "sharkdp",
			"repository": "hexyl",
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
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
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
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
			"linux_asset": "linux64.tar.gz",
			"windows_asset": "win64.zip"
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
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
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
			"linux_asset": "x86_64-unknown-linux-musl",
			"windows_asset": ""
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
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip"
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
			"linux_asset": "tealdeer-linux-x86_64-musl",
			"windows_asset": "windows-x86_64-msvc.exe"
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
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.exe"
		}
	}
}`

func writeDefaultConfiguration(path *string) error {
	filePath := replaceTildePath(*path)
	dirName := filepath.Dir(filePath)

	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}

	_, err = os.Stat(filePath)
	if err == nil {
		fmt.Print("A file already exists at that location. Overwrite? [y/N]")
		var input string
		fmt.Scan(&input)
		if input != "" && (input[0] == 121 || input[0] == 89) {
			return os.WriteFile(filePath, []byte(defaultConfiguration), 0644)
		}

		return nil
	} else {
		return os.WriteFile(filePath, []byte(defaultConfiguration), 0644)
	}
}
