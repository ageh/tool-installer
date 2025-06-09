// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

type App struct {
	downloader     Downloader
	config         Configuration
	cache          Cache
	configLocation string
}

func newApp(configPath string, timeout int) (App, error) {
	var result App

	config, err := readConfiguration(configPath)
	if err != nil {
		return result, fmt.Errorf("could not read configuration: %w", err)
	}

	result.config = config

	cache, err := getCache()
	if err != nil {
		return result, fmt.Errorf("could not obtain cache: %w", err)
	}

	result.cache = cache

	result.downloader = newDownloader(timeout)

	result.configLocation = configPath

	return result, nil
}

func (app *App) addTool() error {
	name := promptNonEmpty("Please enter the name of the tool: ")

	_, found := app.config.Tools[name]
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

	app.config.Tools[name] = Tool{
		Binaries:     binaries,
		Owner:        owner,
		Repository:   repo,
		LinuxAsset:   linux,
		WindowsAsset: windows,
		Description:  description,
	}

	return app.config.save(app.configLocation, false)
}

func (app *App) checkToolVersions(checkAll bool) error {
	results, err := app.getOutdatedTools(checkAll)
	if err != nil {
		return fmt.Errorf("error during check for outdated versions: %w", err)
	}

	table := newTableBuilder([]string{"Name", "Installed", "Available"})

	if len(results) == 0 {
		fmt.Println("All tools are up to date.")
		return nil
	}

	for _, e := range results {
		table.addRow([]string{e.Name, e.Installed, e.Available})
	}

	fmt.Print(table.build())

	return nil
}

func (app *App) installTools(tools []string) error {
	toolDirectory := replaceTildePath(app.config.InstallationDirectory)
	err := makeOutputDirectory(toolDirectory)
	if err != nil {
		return err
	}

	var toInstall map[string]Tool

	if len(tools) > 0 {
		toInstall = make(map[string]Tool, len(tools))
		for _, name := range tools {
			tool, found := app.config.Tools[name]
			if !found {
				fmt.Printf("Error: tool '%s' not found in configuration\n", name)
				continue
			}

			toInstall[name] = tool
		}
	} else {
		toInstall = app.config.Tools
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersionInfo, len(toInstall))

	for name, tool := range toInstall {
		wg.Add(1)
		go func() {
			defer wg.Done()

			currentVersion := app.cache.Tools[name]

			result, err := app.downloader.downloadTool(tool, currentVersion)
			if err != nil {
				fmt.Printf("Error: failed to download '%s', %v\n", name, err)
				return
			}

			if result.updated {
				fmt.Printf("Info: skipping download for '%s' because it is already installed and up to date\n", name)
				return
			}

			err = extractFiles(result.data, result.assetName, tool.Binaries, toolDirectory)
			if err != nil {
				fmt.Printf("Error: failed to extract files for '%s': %v\n", name, err)
				return
			}

			results <- ToolVersionInfo{Name: name, Installed: result.tagName}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		app.cache.add(result.Name, result.Installed)
	}

	err = app.cache.writeCache()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) listTools(longList bool) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	tmp := make([]ToolInfo, len(app.config.Tools))

	i := 0
	for k, v := range app.config.Tools {
		tmp[i] = ToolInfo{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := cache.Tools[k]; found {
			tmp[i].Version = version
		}

		i++
	}

	sort.Sort(ByName[ToolInfo]{tmp})

	var builder TableBuilder

	if longList {
		builder = newTableBuilder([]string{"Name", "Source", "Version", "Description"})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Link, row.Version, row.Description})
		}
	} else {
		builder = newTableBuilderWithLimits([]string{"Name", "Version", "Description"}, map[int]int{2: 50})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Version, row.Description})
		}
	}

	fmt.Print(builder.build())

	return nil
}

func (app *App) removeTools(tools []string) error {
	toolDirectory := replaceTildePath(app.config.InstallationDirectory)

	for _, name := range tools {

		tool, found := app.config.Tools[name]
		if !found {
			continue
		}

		for _, binary := range tool.Binaries {
			n := binary.Name
			if binary.RenameTo != "" {
				n = binary.RenameTo
			}

			path := filepath.Join(toolDirectory, n)
			err := os.Remove(path)
			if err != nil {
				fmt.Printf("Failed to remove binary '%s' for tool '%s'.", n, name)
			}
		}

		delete(app.config.Tools, name)

		app.cache.remove(name)
	}

	err := app.config.save(app.configLocation, false)
	if err != nil {
		return err
	}

	return app.cache.writeCache()
}

func (app *App) updateTools() error {
	outdated, err := app.getOutdatedTools(false)
	if err != nil {
		return err
	}

	tools := make([]string, len(outdated))
	for i, tmp := range outdated {
		tools[i] = tmp.Name
	}

	return app.installTools(tools)
}

func (app *App) toolsFromCache() map[string]Tool {
	tools := make(map[string]Tool, len(app.cache.Tools))
	for name := range app.cache.Tools {
		tools[name] = app.config.Tools[name]
	}

	return tools
}

func (app *App) getOutdatedTools(checkAll bool) ([]ToolVersionInfo, error) {
	var tools map[string]Tool
	if checkAll {
		tools = app.config.Tools
	} else {
		tools = app.toolsFromCache()
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersionInfo, len(tools))

	for name, tool := range tools {
		wg.Add(1)

		go func() {
			defer wg.Done()

			release, err := app.downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error: failed to obtain latest release of tool '%s': %v\n", name, err)
				return
			}

			results <- ToolVersionInfo{Name: name, Installed: app.cache.Tools[name], Available: release.TagName}
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
