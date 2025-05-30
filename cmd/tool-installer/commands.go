// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const addHelp = `Adds a tool to the configuration by prompting the necessary values from the user.

Examples:
tooli add ripgrep
tooli add bat`
const checkHelp = `Checks the configured tools for version updates.

By default only the currently installed tools are check, to change this pass 'all' as an argument to the command.

Examples:

tooli check
tooli check all`
const createConfigHelp = `Creates the default configuration.

By default the configuration is written to '~/.config/tool-installer/config.json',
but this can be changed by passing a different path as an argument.

Examples:

tooli create-config
tooli create-config test.json`
const helpHelp = `Shows the help for the program or command.

Examples:
tooli help
tooli help install`
const installHelp = `Installs tools. If no arguments are provided, it installs all tools in the configuration.
Installs only the named tools if provided with a space separated list of tools to install.

Examples

tooli install
tooli install ripgrep
tooli install ripgrep eza bat fd`
const listHelp = `Lists the tools present in the configuration.

Examples:

tooli list
tooli list long`
const removeHelp = `Removes one or more tools from the configuration.

Examples:
tooli remove ripgrep
tooli remove ripgrep bat micro`
const updateHelp = `Updates all installed tools to their latest version.

Examples:
tooli update`

type TableEntry struct {
	Name        string
	Link        string
	Description string
	Version     string
}

func (t TableEntry) GetName() string {
	return t.Name
}

type VersionTableEntry struct {
	Name      string
	Installed string
	Available string
}

func (v VersionTableEntry) GetName() string {
	return v.Name
}

type Named interface {
	GetName() string
}

// Define a generic type that implements sort.Interface for any slice of Named
type ByName[T Named] struct {
	data []T
}

func (array ByName[T]) Len() int {
	return len(array.data)
}

func (array ByName[T]) Less(i int, j int) bool {
	return array.data[i].GetName() < array.data[j].GetName()
}

func (array ByName[T]) Swap(i int, j int) {
	array.data[i], array.data[j] = array.data[j], array.data[i]
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

func getCommandHelp(command string) string {
	switch command {
	case "a", "add":
		return addHelp
	case "c", "check":
		return checkHelp
	case "cc", "create-config":
		return createConfigHelp
	case "h", "help":
		return helpHelp
	case "i", "install":
		return installHelp
	case "l", "list":
		return listHelp
	case "r", "remove":
		return removeHelp
	case "u", "update":
		return updateHelp
	default:
		return fmt.Sprintf("Error: '%s' is not a valid command", command)
	}
}

func getOutdatedTools(config Configuration, checkAll bool, downloadTimeout int, cache Cache) ([]VersionTableEntry, error) {
	downloader := newDownloader(downloadTimeout)

	var tools map[string]Tool
	if checkAll {
		tools = config.Tools
	} else {
		tools = make(map[string]Tool, len(cache.Tools))
		for name := range cache.Tools {
			tools[name] = config.Tools[name]
		}
	}

	var wg sync.WaitGroup

	results := make(chan VersionTableEntry, len(tools))

	for name, tool := range tools {
		wg.Add(1)

		go func() {
			defer wg.Done()

			release, err := downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error: failed to obtain latest release of tool '%s': %v\n", name, err)
				return
			}

			results <- VersionTableEntry{Name: name, Installed: cache.Tools[name], Available: release.TagName}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	result := make([]VersionTableEntry, 0)

	for r := range results {
		if r.Installed != r.Available {
			result = append(result, r)
		}
	}

	sort.Sort(ByName[VersionTableEntry]{result})

	return result, nil
}

func checkToolVersions(config Configuration, checkAll bool, downloadTimeout int) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	results, err := getOutdatedTools(config, checkAll, downloadTimeout, cache)
	if err != nil {
		return err
	}

	nameSize := 4
	installedSize := 9
	availableSize := 9

	for _, r := range results {
		nameSize = max(nameSize, len(r.Name))
		installedSize = max(installedSize, len(r.Installed))
		availableSize = max(availableSize, len(r.Available))
	}

	if len(results) > 0 {
		fmt.Printf("%-*s    %-*s    %-*s\n\n", nameSize, "Name", installedSize, "Installed", availableSize, "Available")

		for _, e := range results {
			fmt.Printf("%-*s    %-*s    %-*s\n", nameSize, e.Name, installedSize, e.Installed, availableSize, e.Available)
		}
	} else {
		fmt.Println("All tools are up to date.")
	}

	return nil
}

func listTools(config Configuration, longList bool) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	// Minimum sizes based on header line
	nameSize := 4
	linkSize := 16
	descriptionSize := 11
	versionSize := 7

	tmp := make([]TableEntry, len(config.Tools))

	i := 0
	for k, v := range config.Tools {
		tmp[i] = TableEntry{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := cache.Tools[k]; found {
			tmp[i].Version = version
		}

		nameSize = max(nameSize, len(k))
		linkSize = max(linkSize, len(tmp[i].Link))
		descriptionSize = max(descriptionSize, len(v.Description))
		versionSize = max(versionSize, len(tmp[i].Version))

		i++
	}

	sort.Sort(ByName[TableEntry]{tmp})

	if longList {
		fmt.Printf("%-*s    %-*s    %-*s    %-*s\n\n", nameSize, "Name", linkSize, "Owner/Repository", descriptionSize, "Description", versionSize, "Version")

		for _, j := range tmp {
			fmt.Printf("%-*s    %-*s    %-*s    %-*s\n", nameSize, j.Name, linkSize, j.Link, descriptionSize, j.Description, versionSize, j.Version)
		}
	} else {
		descriptionSize = min(descriptionSize, maxShortListDescriptionLength)
		fmt.Printf("%-*s    %-*s       %-*s\n\n", nameSize, "Name", descriptionSize, "Description", versionSize, "Version")

		for _, j := range tmp {
			extra := "   "
			if len(j.Description) > maxShortListDescriptionLength {
				extra = "..."
				j.Description = j.Description[:maxShortListDescriptionLength]
			}
			fmt.Printf("%-*s    %-*s%s    %-*s\n", nameSize, j.Name, descriptionSize, j.Description, extra, versionSize, j.Version)
		}
	}

	return nil
}

func makeOutputDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory ('%s'): %w", path, err)
	}

	return nil
}

func installFiles(binaries []Binary, result DownloadResult, installationDirectory string) (string, error) {
	err := extractFiles(result.data, result.assetName, binaries, installationDirectory)
	if err != nil {
		return "", fmt.Errorf("error during tool installation: %w", err)
	}

	return result.tagName, nil
}

type ToolVersion struct {
	name    string
	version string
}

func installTools(config Configuration, tools []string, downloadTimeout int) error {
	err := makeOutputDirectory(config.InstallationDirectory)
	if err != nil {
		return err
	}

	cache, err := getCache()
	if err != nil {
		return err
	}

	downloader := newDownloader(downloadTimeout)

	var toInstall map[string]Tool

	if len(tools) > 0 {
		toInstall = make(map[string]Tool, len(tools))
		for _, name := range tools {
			tool, found := config.Tools[name]
			if !found {
				fmt.Printf("Error: tool '%s' not found in configuration\n", name)
				continue
			}

			toInstall[name] = tool
		}
	} else {
		toInstall = config.Tools
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersion, len(toInstall))

	for name, tool := range toInstall {
		wg.Add(1)
		go func() {
			defer wg.Done()

			currentVersion := cache.Tools[name]

			result, err := downloader.downloadTool(tool, currentVersion)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if result.updated {
				fmt.Printf("Info: skipping download for '%v' because it is already installed and up to date\n", name)
				return
			}

			installedVersion, err := installFiles(tool.Binaries, result, config.InstallationDirectory)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			results <- ToolVersion{name: name, version: installedVersion}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		cache.add(result.name, result.version)
	}

	err = cache.writeCache()
	if err != nil {
		return err
	}

	return nil
}

func updateTools(config Configuration, downloadTimeout int) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	outdated, err := getOutdatedTools(config, false, downloadTimeout, cache)
	if err != nil {
		return err
	}

	tools := make([]string, len(outdated))
	for i, tmp := range outdated {
		tools[i] = tmp.Name
	}

	return installTools(config, tools, downloadTimeout)
}

func prompt(text string) string {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}

	return strings.TrimSpace(input)
}

func promptNonEmpty(text string) string {
	fmt.Print(text)

	for {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return ""
		}

		result := strings.TrimSpace(input)

		if result != "" {
			return result
		}

		fmt.Print("Input must not be empty. Please try again: ")
	}
}

func promptForBinary() (Binary, bool) {
	binary := prompt("Binary name: ")
	rename := prompt("Rename binary to (leave empty if no rename): ")

	if binary == "" {
		return Binary{}, false
	}

	return Binary{Name: binary, RenameTo: rename}, true
}

func addTool(config *Configuration, name string, configPath string) error {
	_, found := config.Tools[name]
	if found {
		return errors.New("an entry for this tool already exists. If you want to modify it, please edit the configuration file")
	}

	fmt.Printf("Creating configuration entry for %s:\n", name)

	description := promptNonEmpty("Short description: ")
	owner := promptNonEmpty("GitHub user/org: ")
	repo := promptNonEmpty("Repository name: ")

	windows := prompt("Windows asset name: ")
	linux := prompt("Linux asset name: ")

	prefix := prompt("Asset prefix (leave empty if not needed): ")

	binary := promptNonEmpty("Binary name: ")
	rename := prompt("Rename binary to (leave empty if no rename): ")

	furtherEntries := prompt("Does this tool have more binaries? [y/N]: ")

	binaries := []Binary{{Name: binary, RenameTo: rename}}

	if furtherEntries == "y" || furtherEntries == "Y" {
		for {
			binary, ok := promptForBinary()

			if !ok {
				break
			}

			binaries = append(binaries, binary)
		}
	}

	config.Tools[name] = Tool{
		Binaries:     binaries,
		Owner:        owner,
		Repository:   repo,
		LinuxAsset:   linux,
		WindowsAsset: windows,
		AssetPrefix:  prefix,
		Description:  description,
	}

	return config.save(configPath, false)
}

func removeTools(config *Configuration, tools []string, configPath string) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	dir := config.InstallationDirectory

	for _, name := range tools {

		tool, found := config.Tools[name]
		if !found {
			continue
		}

		for _, binary := range tool.Binaries {
			n := binary.Name
			if binary.RenameTo != "" {
				n = binary.RenameTo
			}

			path := filepath.Join(dir, n)
			err = os.Remove(path)
			if err != nil {
				fmt.Printf("Failed to remove binary '%s' for tool '%s'.", n, name)
			}
		}

		delete(config.Tools, name)

		cache.remove(name)
	}

	err = config.save(configPath, false)
	if err != nil {
		return err
	}

	return cache.writeCache()
}
