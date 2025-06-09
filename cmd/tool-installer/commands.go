// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type ToolInfo struct {
	Name        string
	Link        string
	Description string
	Version     string
}

func (t ToolInfo) GetName() string {
	return t.Name
}

type ToolVersionInfo struct {
	Name      string
	Installed string
	Available string
}

func (v ToolVersionInfo) GetName() string {
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

func getOutdatedTools(config Configuration, checkAll bool, downloadTimeout int, cache Cache) ([]ToolVersionInfo, error) {
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

	results := make(chan ToolVersionInfo, len(tools))

	for name, tool := range tools {
		wg.Add(1)

		go func() {
			defer wg.Done()

			release, err := downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error: failed to obtain latest release of tool '%s': %v\n", name, err)
				return
			}

			results <- ToolVersionInfo{Name: name, Installed: cache.Tools[name], Available: release.TagName}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	result := make([]ToolVersionInfo, 0)

	for r := range results {
		if r.Installed != r.Available {
			result = append(result, r)
		}
	}

	sort.Sort(ByName[ToolVersionInfo]{result})

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

	table := newTableBuilder([]string{"Name", "Installed", "Available"})

	if len(results) > 0 {
		for _, e := range results {
			table.addRow([]string{e.Name, e.Installed, e.Available})
		}

		fmt.Print(table.build())
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

	tmp := make([]ToolInfo, len(config.Tools))

	i := 0
	for k, v := range config.Tools {
		tmp[i] = ToolInfo{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := cache.Tools[k]; found {
			tmp[i].Version = version
		}

		i++
	}

	sort.Sort(ByName[ToolInfo]{tmp})

	if longList {
		builder := newTableBuilder([]string{"Name", "Source", "Version", "Description"})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Link, row.Version, row.Description})
		}

		fmt.Print(builder.build())
	} else {
		builder := newTableBuilderWithLimits([]string{"Name", "Version", "Description"}, map[int]int{2: 50})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Version, row.Description})
		}

		fmt.Print(builder.build())
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

	results := make(chan ToolVersionInfo, len(toInstall))

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

			results <- ToolVersionInfo{Name: name, Installed: installedVersion}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		cache.add(result.Name, result.Installed)
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
	reader := bufio.NewReader(os.Stdin)

	for {
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

func promptRegex(text string) string {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return ""
		}

		result := strings.TrimSpace(input)

		_, err = regexp.Compile(result)
		if err == nil {
			return result
		}

		fmt.Print("Input must be a valid regular expression. Please try again: ")
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

	windows := promptRegex("Windows asset name (regex): ")
	linux := promptRegex("Linux asset name (regex): ")

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
